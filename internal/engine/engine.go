package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/baml_client/types"
	"github.com/lixianmin/pc/internal/agent_tools"
	"github.com/lixianmin/pc/internal/debug"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/internal/skill"
	"github.com/lixianmin/pc/internal/task"
	"github.com/lixianmin/pc/pkg/ks"
	pkgtypes "github.com/lixianmin/pc/pkg/types"
)

type PluginCaller interface {
	ListPlugins() []*pkgtypes.Plugin
	CallPlugin(plugin *pkgtypes.Plugin, method string, params any) (any, error)
}

// Engine is the core engine implementation.
type Engine struct {
	pluginManager  *plugin.PluginManager
	sessions       map[string]*Session
	mu             sync.RWMutex
	llmPlugin      *pkgtypes.Plugin
	systemPrompt   string
	maxIterations  int
	toolTimeout    time.Duration
	promptRecorder *debug.PromptRecorder

	// Task management
	taskManager *task.Manager

	// Skill management
	skillManager *skill.SkillManager

	llmClient *Client
}

// NewEngine creates a new core engine.
func NewEngine(pm *plugin.PluginManager) *Engine {
	return &Engine{
		pluginManager: pm,
		sessions:      make(map[string]*Session),
		maxIterations: 10,
		toolTimeout:   30 * time.Second,
		taskManager:   task.NewManager(""),
		llmClient:     NewClient(),
	}
}

// SetPromptRecorder sets the prompt recorder for debugging.
func (my *Engine) SetPromptRecorder(recorder *debug.PromptRecorder) {
	my.promptRecorder = recorder
}

// SetToolTimeout sets the tool execution timeout.
func (my *Engine) SetToolTimeout(d time.Duration) {
	my.toolTimeout = d
}

// SetLLMPlugin sets the LLM plugin to use for generating responses.
func (my *Engine) SetLLMPlugin(p *pkgtypes.Plugin) {
	my.llmPlugin = p
}

// SetSystemPrompt sets the complete system prompt for LLM calls.
func (my *Engine) SetSystemPrompt(prompt string) {
	my.systemPrompt = prompt
}

// BuildSystemPrompt builds the dynamic system prompt with tools and skills.
func (my *Engine) BuildSystemPrompt() string {
	return my.buildSystemPrompt()
}

// GetTaskManager returns the task manager.
func (my *Engine) GetTaskManager() *task.Manager {
	return my.taskManager
}

// SetSkillDir sets the skill directory and loads skills.
func (my *Engine) SetSkillDir(dir string) error {
	my.skillManager = skill.NewSkillManager()
	_, err := my.skillManager.LoadSkills(dir)
	return err
}

// ListSkills returns all loaded skills.
func (my *Engine) ListSkills() []skill.Skill {
	if my.skillManager == nil {
		return nil
	}
	return my.skillManager.ListSkills()
}

// GetSkill returns a skill by name.
func (my *Engine) GetSkill(name string) *skill.Skill {
	return my.skillManager.GetSkill(name)
}

// ProcessMessage processes an incoming message with ReAct loop.
func (my *Engine) ProcessMessage(ctx context.Context, sessionId, message string) (string, error) {
	if sessionId == "" {
		return "", ks.NewAppError("EngineProcess", "session ID cannot be empty")
	}

	if message == "" {
		return "", ks.NewAppError("EngineProcess", "message cannot be empty")
	}

	my.mu.RLock()
	session, exists := my.sessions[sessionId]
	my.mu.RUnlock()

	if !exists {
		return "", ks.NewAppErrorf("EngineProcess", "session not found: %s", sessionId)
	}

	logo.Info("[Session:", sessionId, "] User:", message)

	if err := session.AddMessage("user", message); err != nil {
		return "", ks.WrapAppError("EngineProcess", "failed to save user message", err)
	}

	var response string
	hasLLM := my.llmClient != nil || (my.pluginManager != nil && my.llmPlugin != nil)
	if hasLLM {
		logo.Info("[Session:", sessionId, "] Starting ReAct loop")
		resp, err := my.reactLoop(ctx, session, message)
		if err != nil {
			logo.Error("[Session:", sessionId, "] ReAct loop failed:", err)
			return "", ks.WrapAppError("EngineProcess", "failed to process message", err)
		}
		response = resp
		logo.Info("[Session:", sessionId, "] ReAct loop completed, response length:", len(response))
	} else {
		response = fmt.Sprintf("Echo: %s", message)
		logo.Info("[Session:", sessionId, "] No LLM plugin, echoing")
	}

	if err := session.AddMessage("assistant", response); err != nil {
		return "", ks.WrapAppError("EngineProcess", "failed to save assistant response", err)
	}

	logResp := response
	if len(logResp) > 100 {
		logResp = logResp[:97] + "..."
	}
	logo.Info("[Session:", sessionId, "] Assistant:", logResp)

	return response, nil
}

// reactLoop implements the ReAct (Reasoning + Acting) loop.
func (my *Engine) reactLoop(ctx context.Context, session *Session, initialMessage string) (string, error) {
	currentMessage := initialMessage

	for iteration := 0; iteration < my.maxIterations; iteration++ {
		logo.JsonI("iteration", iteration+1)

		var result, err = my.callLLM(ctx, session, currentMessage)
		if err != nil {
			return "", ks.WrapAppError("EngineReact", fmt.Sprintf("LLM call failed at iteration %d", iteration+1), err)
		}

		var chatResponse = result.AsChatResponse()
		if chatResponse != nil {
			logo.Info("[ReAct] ChatResponse received, returning final response")
			return chatResponse.Content, nil
		}

		logo.Info("[ReAct] Tool call detected")

		if err := session.AddMessage("assistant", "[Tool call]"); err != nil {
			return "", ks.WrapAppError("EngineReact", "failed to save assistant message", err)
		}

		// 执行工具调用
		var toolCtx, cancel = context.WithTimeout(ctx, my.toolTimeout)
		defer cancel()

		var toolResult, toolErr = my.callTool(toolCtx, result)
		if toolErr != nil {
			logo.Error("[ReAct] Tool execution failed:", toolErr)
			return "", ks.WrapAppError("EngineReact", "tool execution failed", toolErr)
		}

		if err := session.AddMessage("system", toolResult); err != nil {
			return "", ks.WrapAppError("EngineReact", "failed to save tool results", err)
		}

		// 这里是有用的， 必须重设一下
		// currentMessage = toolResultsMessage
	}

	return "", ks.NewAppErrorf("EngineReact", "exceeded maximum iterations (%d)", my.maxIterations)
}

// callLLM calls the LLM via BAML to generate a response.
func (my *Engine) callLLM(ctx context.Context, session *Session, message string) (*ChatResult, error) {
	startTime := time.Now()

	var systemPrompt = my.buildSystemPrompt()
	history := session.GetMessages()
	var messages []types.Message
	for _, msg := range history {
		messages = append(messages, types.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, types.Message{
		Role:    "user",
		Content: message,
	})

	result, err := my.llmClient.Chat(ctx, messages, systemPrompt)
	if err != nil {
		logo.Error("[Engine.callLLM] Chat failed:", err)
		return nil, err
	}

	elapsed := time.Since(startTime)
	var totalChars int
	for _, m := range messages {
		totalChars += len(m.Content)
	}
	logo.Info("[Engine.callLLM] LLM call completed, chars=", totalChars, " time=", elapsed)

	return &result, nil
}

func (my *Engine) callTool(ctx context.Context, result *ChatResult) (string, error) {
	if item := result.AsBashTool(); item != nil {
		return agent_tools.Bash(ctx, "", item.Command)
	}

	if item := result.AsEditTool(); item != nil {
		return agent_tools.Edit(ctx, item.FilePath, item.OldString, item.NewString, false)
	}

	if item := result.AsReadTool(); item != nil {
		return agent_tools.Read(ctx, item.FilePath, 0, 100)
	}

	if item := result.AsWriteTool(); item != nil {
		return agent_tools.Write(ctx, item.FilePath, item.Content)
	}

	return "", ks.NewAppError("EngineCallTool", "unsupported tool type")
}

func (my *Engine) buildSystemPrompt() string {
	var parts []string

	if my.systemPrompt != "" {
		parts = append(parts, my.systemPrompt)
	}

	var skills = my.ListSkills()
	if len(skills) > 0 {
		parts = append(parts, "", my.buildSkillGuide(skills))
	}

	return strings.Join(parts, "\n")
}

func (my *Engine) buildSkillGuide(skills []skill.Skill) string {
	var parts []string

	parts = append(parts, "## 技能使用指南")
	parts = append(parts, "")
	parts = append(parts, "技能是由多个步骤组成的复合能力，当需要执行复杂任务时可以参考。")
	parts = append(parts, "")

	parts = append(parts, "### 可用技能")
	for _, s := range skills {
		description := s.Description
		if s.Description == "" {
			description = "复合任务"
		}
		parts = append(parts, fmt.Sprintf("- **%s**: %s", s.Name, description))
	}

	parts = append(parts, "")
	parts = append(parts, "### 使用方式")
	parts = append(parts, "当需要执行复杂任务时，可以参考技能中的步骤来指导你的操作。")

	return strings.Join(parts, "\n")
}

// FetchSession returns the session if exists, otherwise creates a new one.
func (my *Engine) FetchSession(sessionId string) *Session {
	if sessionId == "" {
		return nil
	}

	my.mu.Lock()
	defer my.mu.Unlock()

	session, exists := my.sessions[sessionId]
	if !exists {
		session = NewSession(sessionId)
		my.sessions[sessionId] = session
	}
	return session
}

// Close closes the engine.
func (my *Engine) Close() error {
	my.mu.Lock()
	defer my.mu.Unlock()
	for key := range my.sessions {
		delete(my.sessions, key)
	}
	return nil
}
