package task

import (
	"fmt"
	"strings"
)

// LLMClient is the interface for LLM operations
type LLMClient interface {
	DecomposeGoal(goal string) ([]string, error)
}

// Decomposer handles goal decomposition into tasks
type Decomposer struct {
	llm LLMClient
}

// NewDecomposer creates a new goal decomposer
func NewDecomposer() *Decomposer {
	return &Decomposer{}
}

// SetLLMClient sets the LLM client for decomposition
func (my *Decomposer) SetLLMClient(client LLMClient) {
	my.llm = client
}

// ParseGoal parses user input to extract the goal
func (my *Decomposer) ParseGoal(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("goal cannot be empty")
	}

	// Remove common prefixes
	prefixes := []string{
		"我的目标是：",
		"我的目标是:",
		"目标是：",
		"目标是:",
		"goal:",
		"goal：",
		"我想",
		"我要",
		"请帮我",
	}

	goal := input
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(goal), strings.ToLower(prefix)) {
			goal = strings.TrimSpace(goal[len(prefix):])
		}
	}

	if goal == "" {
		return "", fmt.Errorf("goal is empty after parsing")
	}

	return goal, nil
}

// Decompose breaks down a goal into steps using LLM
func (my *Decomposer) Decompose(goal string) ([]string, error) {
	if goal == "" {
		return nil, fmt.Errorf("goal cannot be empty")
	}

	// If LLM client is set, use it
	if my.llm != nil {
		return my.llm.DecomposeGoal(goal)
	}

	// Default decomposition for simple goals
	// This is a fallback when no LLM is available
	return my.simpleDecomposition(goal), nil
}

// simpleDecomposition provides basic decomposition without LLM
func (my *Decomposer) simpleDecomposition(goal string) []string {
	// Basic rule-based decomposition
	goalLower := strings.ToLower(goal)

	// Common patterns
	if strings.Contains(goalLower, "deploy") {
		return []string{
			"Prepare deployment environment",
			"Run pre-deployment checks",
			"Execute deployment",
			"Verify deployment success",
		}
	}

	if strings.Contains(goalLower, "build") || strings.Contains(goalLower, "develop") {
		return []string{
			"Design architecture",
			"Implement core features",
			"Write tests",
			"Review and refactor",
		}
	}

	if strings.Contains(goalLower, "test") {
		return []string{
			"Write test cases",
			"Set up test environment",
			"Execute tests",
			"Analyze results",
		}
	}

	// Default generic steps
	return []string{
		"Analyze requirements",
		"Plan implementation",
		"Execute plan",
		"Review and finalize",
	}
}

// ConvertToTask converts a goal and steps into a Task
func (my *Decomposer) ConvertToTask(goal string, steps []string) (*Task, error) {
	task, err := NewTask(goal)
	if err != nil {
		return nil, err
	}

	for _, step := range steps {
		if err := task.AddStep(step); err != nil {
			return nil, fmt.Errorf("failed to add step '%s': %w", step, err)
		}
	}

	return task, nil
}

// DecomposeAndCreateTask parses input, decomposes goal, and creates a task
func (my *Decomposer) DecomposeAndCreateTask(input string) (*Task, error) {
	// Parse goal from input
	goal, err := my.ParseGoal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse goal: %w", err)
	}

	// Decompose goal into steps
	steps, err := my.Decompose(goal)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose goal: %w", err)
	}

	// Convert to task
	task, err := my.ConvertToTask(goal, steps)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}
