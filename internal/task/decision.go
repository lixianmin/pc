package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ActionType represents the type of action to take
type ActionType string

const (
	// ActionRespond indicates direct response to user
	ActionRespond ActionType = "respond"
	// ActionUseTool indicates using a tool
	ActionUseTool ActionType = "use_tool"
	// ActionUseSkill indicates using a skill
	ActionUseSkill ActionType = "use_skill"
	// ActionWait indicates waiting for clarification
	ActionWait ActionType = "wait"
)

// Message represents a message in the conversation history
type Message struct {
	Role    string
	Content string
}

// Context represents the decision context
type Context struct {
	UserMessage string
	History     []Message
}

// Decision represents a decision made by the engine
type Decision struct {
	Action     ActionType
	Tool       string
	Skill      string
	Reason     string
	Confidence float64
}

// LLMCompletionClient is the interface for LLM completion operations
type LLMCompletionClient interface {
	// Complete sends a prompt to the LLM and returns the response
	Complete(ctx context.Context, prompt string) (string, error)
}

// DecisionEngine handles autonomous decision making
type DecisionEngine struct {
	toolKeywords  map[string][]string
	skillKeywords map[string][]string
	llmClient     LLMCompletionClient
	useLLM        bool
}

// NewDecisionEngine creates a new decision engine
func NewDecisionEngine() *DecisionEngine {
	return &DecisionEngine{
		toolKeywords: map[string][]string{
			"weather": {"weather", "temperature", "forecast", "rain", "sunny"},
			"search":  {"search", "find", "look up", "google", "query"},
			"file":    {"read", "write", "file", "config", "document"},
			"time":    {"time", "date", "current", "now"},
		},
		skillKeywords: map[string][]string{
			"deploy":  {"deploy", "release", "publish", "ship"},
			"test":    {"test", "check", "verify", "validate"},
			"build":   {"build", "compile", "make", "package"},
			"analyze": {"analyze", "review", "audit", "inspect"},
		},
		useLLM: false, // Disabled by default until LLM client is set
	}
}

// SetLLMClient sets the LLM client for LLM-based decision making
func (my *DecisionEngine) SetLLMClient(client LLMCompletionClient) {
	my.llmClient = client
	my.useLLM = client != nil
}

// Decide makes a decision based on context
func (my *DecisionEngine) Decide(ctx Context) (Decision, error) {
	// Check for empty message
	if strings.TrimSpace(ctx.UserMessage) == "" {
		return Decision{
			Action: ActionWait,
			Reason: "Empty user message",
		}, nil
	}

	// Use LLM for decision if available
	if my.useLLM && my.llmClient != nil {
		return my.decideWithLLM(context.Background(), ctx)
	}

	// Fall back to keyword-based decision
	return my.decideWithKeywords(ctx)
}

// decideWithLLM uses LLM to make a decision
func (my *DecisionEngine) decideWithLLM(ctx context.Context, dctx Context) (Decision, error) {
	prompt := my.buildDecisionPrompt(dctx)

	response, err := my.llmClient.Complete(ctx, prompt)
	if err != nil {
		// Fall back to keyword matching if LLM fails
		return my.decideWithKeywords(dctx)
	}

	// Parse LLM response
	return my.parseLLMResponse(response, dctx)
}

// buildDecisionPrompt builds the prompt for LLM decision making
func (my *DecisionEngine) buildDecisionPrompt(ctx Context) string {
	// Build conversation context
	var historyStr string
	for _, msg := range ctx.History {
		historyStr += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	// Build available tools and skills
	var tools []string
	for tool := range my.toolKeywords {
		tools = append(tools, tool)
	}
	var skills []string
	for skill := range my.skillKeywords {
		skills = append(skills, skill)
	}

	prompt := fmt.Sprintf(`You are an AI assistant that decides how to handle user requests.

Available tools: %v
Available skills: %v

Conversation history:
%s
Current user message: %s

Respond in JSON format:
{
  "action": "respond|use_tool|use_skill|wait",
  "tool": "tool_name_if_action_is_use_tool",
  "skill": "skill_name_if_action_is_use_skill",
  "reason": "explanation of the decision",
  "confidence": 0.85
}

Guidelines:
- Use "respond" for general questions, greetings, or when no tool/skill is needed
- Use "use_tool" for specific tool requests (weather, search, file operations, time)
- Use "use_skill" for complex workflows (deploy, test, build, analyze)
- Use "wait" if the request is unclear or needs clarification
- Confidence should be between 0.0 and 1.0`,
		tools, skills, historyStr, ctx.UserMessage)

	return prompt
}

// parseLLMResponse parses the LLM response into a Decision
func (my *DecisionEngine) parseLLMResponse(response string, ctx Context) (Decision, error) {
	// Try to extract JSON from response (handle markdown code blocks)
	jsonStr := extractJSON(response)

	var result struct {
		Action     string  `json:"action"`
		Tool       string  `json:"tool"`
		Skill      string  `json:"skill"`
		Reason     string  `json:"reason"`
		Confidence float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Fall back to keyword matching if parsing fails
		return my.decideWithKeywords(ctx)
	}

	decision := Decision{
		Action:     ActionType(result.Action),
		Tool:       result.Tool,
		Skill:      result.Skill,
		Reason:     result.Reason,
		Confidence: result.Confidence,
	}

	// Validate action type
	switch decision.Action {
	case ActionRespond, ActionUseTool, ActionUseSkill, ActionWait:
		// Valid
	default:
		// Invalid action, fall back to respond
		decision.Action = ActionRespond
		decision.Reason = "Invalid action from LLM, defaulting to respond"
	}

	return decision, nil
}

// extractJSON extracts JSON from a string (handles markdown code blocks)
func extractJSON(s string) string {
	// Try to find JSON in code blocks
	if idx := strings.Index(s, "```json"); idx != -1 {
		start := idx + 7
		if end := strings.Index(s[start:], "```"); end != -1 {
			return strings.TrimSpace(s[start : start+end])
		}
	}

	// Try to find JSON between braces
	if idx := strings.Index(s, "{"); idx != -1 {
		if end := strings.LastIndex(s, "}"); end != -1 && end > idx {
			return s[idx : end+1]
		}
	}

	return s
}

// decideWithKeywords makes a decision using keyword matching (fallback)
func (my *DecisionEngine) decideWithKeywords(ctx Context) (Decision, error) {
	// Check if skill is needed
	if skill, found := my.SelectSkill(ctx.UserMessage); found {
		return Decision{
			Action: ActionUseSkill,
			Skill:  skill,
			Reason: "Skill match found: " + skill,
		}, nil
	}

	// Check if tool is needed
	if need, tool := my.NeedsTool(ctx.UserMessage); need {
		return Decision{
			Action: ActionUseTool,
			Tool:   tool,
			Reason: "Tool match found: " + tool,
		}, nil
	}

	// Default to direct response
	return Decision{
		Action: ActionRespond,
		Reason: "No tool or skill match, respond directly",
	}, nil
}

// NeedsTool determines if a tool is needed based on the message
func (my *DecisionEngine) NeedsTool(message string) (bool, string) {
	messageLower := strings.ToLower(message)

	for tool, keywords := range my.toolKeywords {
		for _, keyword := range keywords {
			if strings.Contains(messageLower, keyword) {
				return true, tool
			}
		}
	}

	return false, ""
}

// SelectSkill selects the appropriate skill for the message
func (my *DecisionEngine) SelectSkill(message string) (string, bool) {
	messageLower := strings.ToLower(message)

	for skill, keywords := range my.skillKeywords {
		for _, keyword := range keywords {
			if strings.Contains(messageLower, keyword) {
				return skill, true
			}
		}
	}

	return "", false
}

// CalculateConfidence calculates the confidence score for a decision
func (my *DecisionEngine) CalculateConfidence(ctx Context, decision Decision) float64 {
	// Empty message has low confidence
	if strings.TrimSpace(ctx.UserMessage) == "" {
		return 0.3
	}

	// Direct response has high confidence for simple greetings
	if decision.Action == ActionRespond {
		greetings := []string{"hello", "hi", "hey", "good morning", "good afternoon", "good evening"}
		msgLower := strings.ToLower(ctx.UserMessage)
		for _, greeting := range greetings {
			if strings.Contains(msgLower, greeting) {
				return 0.9
			}
		}
	}

	// Tool match confidence
	if decision.Action == ActionUseTool && decision.Tool != "" {
		return 0.85
	}

	// Skill match confidence
	if decision.Action == ActionUseSkill && decision.Skill != "" {
		return 0.8
	}

	return 0.7
}
