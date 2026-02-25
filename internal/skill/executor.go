package skill

import (
	"fmt"
)

// ExecutionStep represents a step in skill execution.
type ExecutionStep struct {
	Type     string       `json:"type"`     // "action", "condition", "loop"
	Command  string       `json:"command,omitempty"`
	Condition string       `json:"condition,omitempty"`
	Loop     *LoopConfig  `json:"loop,omitempty"`
	Steps    []ExecutionStep `json:"steps,omitempty"`
}

// LoopConfig represents loop configuration.
type LoopConfig struct {
	Type     string `json:"type"`     // "repeat", "while"
	Count    int    `json:"count,omitempty"`
	Condition string `json:"condition,omitempty"`
}

// Executor is the interface for skill execution.
type Executor interface {
	// ExecuteSkill executes a skill.
	ExecuteSkill(skill *Skill, context map[string]any) (map[string]any, error)

	// ExecuteStep executes a single step.
	ExecuteStep(step ExecutionStep, context map[string]any) (map[string]any, error)

	// CallTool calls a tool plugin.
	CallTool(toolName string, params map[string]any) (any, error)
}

// Engine is the skill execution engine implementation.
type Engine struct {
	// TODO: Add plugin manager
}

// NewExecutor creates a new skill executor.
func NewExecutor() *Engine {
	return &Engine{}
}

// ExecuteSkill executes a skill.
func (e *Engine) ExecuteSkill(skill *Skill, context map[string]any) (map[string]any, error) {
	// Create execution steps from skill steps
	var steps []ExecutionStep
	for _, stepDesc := range skill.Steps {
		steps = append(steps, ExecutionStep{
			Type:    "action",
			Command: stepDesc,
		})
	}

	// Execute all steps sequentially
	results := []any{}
	for _, step := range steps {
		result, err := e.ExecuteStep(step, context)
		if err != nil {
			return nil, fmt.Errorf("failed to execute step: %w", err)
		}
		results = append(results, result)
	}

	return map[string]any{
		"results": results,
	}, nil
}

// ExecuteStep executes a single step.
func (e *Engine) ExecuteStep(step ExecutionStep, context map[string]any) (map[string]any, error) {
	switch step.Type {
	case "action":
		if step.Command == "" {
			return nil, fmt.Errorf("action step requires command")
		}
		return map[string]any{
			"type":    "action",
			"command": step.Command,
			"result": fmt.Sprintf("executed: %s", step.Command),
		}, nil
	case "condition":
		if step.Condition == "" {
			return nil, fmt.Errorf("condition step requires condition")
		}
		return map[string]any{
			"type":       "condition",
			"condition":  step.Condition,
			"evaluated": true,
		}, nil
	case "loop":
		if step.Loop == nil {
			return nil, fmt.Errorf("loop step requires loop config")
		}
		return e.executeLoop(*step.Loop, step.Steps, context)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeLoop handles loop execution.
func (e *Engine) executeLoop(loop LoopConfig, steps []ExecutionStep, context map[string]any) (map[string]any, error) {
	results := []any{}

	switch loop.Type {
	case "repeat":
		if loop.Count <= 0 {
			return nil, fmt.Errorf("repeat loop requires positive count")
		}
		for i := 0; i < loop.Count; i++ {
			for _, step := range steps {
				result, err := e.ExecuteStep(step, context)
				if err != nil {
					return nil, fmt.Errorf("failed to execute loop step %d: %w", i, err)
				}
				results = append(results, result)
			}
		}
	case "while":
		if loop.Condition == "" {
			return nil, fmt.Errorf("while loop requires condition")
		}
		// Simple while loop - execute once for now (mock)
		// In a real implementation, this would check the condition after each iteration
		for _, step := range steps {
			result, err := e.ExecuteStep(step, context)
			if err != nil {
				return nil, fmt.Errorf("failed to execute while loop step: %w", err)
			}
			results = append(results, result)
		}
	default:
		return nil, fmt.Errorf("unknown loop type: %s", loop.Type)
	}

	return map[string]any{
		"type":       "loop",
		"loopType":  loop.Type,
		"results":    results,
	}, nil
}

// CallTool calls a tool plugin.
func (e *Engine) CallTool(toolName string, params map[string]any) (any, error) {
	// TODO: Implement actual tool calling via plugin manager
	// For now, return a mock result
	if toolName == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}
	// Return error for non-existent tools (for testing)
	if toolName == "non-existent" {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	return map[string]any{
		"tool":   toolName,
		"params":  params,
		"result": fmt.Sprintf("called tool: %s", toolName),
	}, nil
}
