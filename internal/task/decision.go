package task

import (
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
	Action   ActionType
	Tool     string
	Skill    string
	Reason   string
	Confidence float64
}

// DecisionEngine handles autonomous decision making
type DecisionEngine struct {
	toolKeywords  map[string][]string
	skillKeywords map[string][]string
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
			"deploy":   {"deploy", "release", "publish", "ship"},
			"test":     {"test", "check", "verify", "validate"},
			"build":    {"build", "compile", "make", "package"},
			"analyze":  {"analyze", "review", "audit", "inspect"},
		},
	}
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
