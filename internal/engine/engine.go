package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/baml_client"
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
	taskManager  *task.Manager
	skillManager *skill.SkillManager
	Stream       *EngineStream
}

// NewEngine creates a new core engine.
func NewEngine(pm *plugin.PluginManager) *Engine {
	var my = &Engine{
		pluginManager: pm,
		sessions:      make(map[string]*Session),
		maxIterations: 10,
		toolTimeout:   30 * time.Second,
		taskManager:   task.NewManager(""),
	}

	my.Stream = &EngineStream{dad: my}
	return my
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
func (my *Engine) ProcessMessage(ctx context.Context, session *Session, message string) (string, error) {
	if session == nil {
		return "", ks.TraceError("NilSession")
	}

	if message == "" {
		return "", ks.TraceError("EmptyMessage")
	}

	var sessionId = session.Id
	logo.JsonI("sessionId", sessionId, "message", message)

	session.AddMessage("user", message)
	var response, err = my.reactLoop(ctx, session)
	return response, err
}

// reactLoop implements the ReAct (Reasoning + Acting) loop.
func (my *Engine) reactLoop(ctx context.Context, session *Session) (string, error) {
	var sessionId = session.Id
	for iteration := 0; iteration < my.maxIterations; iteration++ {
		logo.JsonI("sessionId", sessionId, "iteration", iteration+1)

		var chatResult, err = my.llmChat(ctx, session)
		if err != nil {
			return "", ks.TraceError("CallLLMFailed", "sessionId", sessionId, "err", err)
		}

		var chatResponse = chatResult.AsChatResponse()
		if chatResponse != nil {
			var response = chatResponse.Content
			session.AddMessage("assistant", response)
			logo.JsonI("sessionId", sessionId, "response", response)

			return response, nil
		}

		logo.JsonI("title", "tool use", "sessionId", sessionId)

		// 使用工具
		var toolCtx, cancel = context.WithTimeout(ctx, my.toolTimeout)
		defer cancel()

		var toolResult, toolErr = my.useTool(toolCtx, chatResult)
		if toolErr != nil {
			// 失败的时候， 把失败信息给llm，要求重试
			session.AddMessage("assistant", toolErr.Error())
			logo.JsonI("sessionId", sessionId, "err", toolErr)
			continue
		}

		session.AddMessage("assistant", toolResult)
	}

	return "", ks.TraceError("IterationExceeded", "maxIterations", my.maxIterations)
}

// llmChat calls the LLM via BAML to generate a response.
func (my *Engine) llmChat(ctx context.Context, session *Session) (*ChatResult, error) {
	var startTime = time.Now()

	var systemPrompt = my.buildSystemPrompt()
	var messages = session.AsBamlMessages()
	my.printPrompt(systemPrompt, messages)

	var chatResult, err = baml_client.Chat(ctx, systemPrompt, messages)
	if err != nil {
		return nil, ks.TraceError("BamlChatError", "err", err)
	}

	var totalChars int
	for _, m := range messages {
		totalChars += len(m.Content)
	}

	logo.JsonI("totalChars", totalChars, "elapsed", time.Since(startTime))
	return &chatResult, nil
}

func (my *Engine) printPrompt(systemPrompt string, messages []types.Message) {
	var sb strings.Builder
	sb.WriteString("=== System Prompt ===\n")
	sb.WriteString(systemPrompt)
	sb.WriteString("\n\n")

	sb.WriteString("=== Conversation History ===\n")
	for _, m := range messages {
		fmt.Fprintf(&sb, "[%s] %s\n", m.Role, m.Content)
	}

	logo.Info(sb.String())
}

func (my *Engine) useTool(ctx context.Context, result *ChatResult) (string, error) {
	if item := result.AsBash(); item != nil {
		var timeout = time.Duration(defaultInt(item.Timeout, 600000) * int(time.Millisecond))
		var toolCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()

		return agent_tools.Bash(toolCtx, item.Command)
	}

	if item := result.AsEdit(); item != nil {
		var expectedReplacements = defaultInt(item.ExpectedReplacements, 1)
		return agent_tools.Edit(ctx, item.FilePath, item.OldString, item.NewString, expectedReplacements)
	}

	if item := result.AsRead(); item != nil {
		var offset = defaultInt(item.Offset, 0)
		var limit = defaultInt(item.Limit, 0)
		return agent_tools.Read(ctx, item.FilePath, offset, limit)
	}

	if item := result.AsWrite(); item != nil {
		return agent_tools.Write(ctx, item.FilePath, item.Content)
	}

	return "", ks.TraceError("UnknownToolType")
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
