package evolution

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// SelfTester handles self-testing and validation
type SelfTester struct {
	projectDir string
}

// TestResult represents a test result
type TestResult struct {
	Passed   bool   `json:"passed"`
	Package  string `json:"package"`
	Tests    int    `json:"tests"`
	Failures int    `json:"failures"`
	Errors   int    `json:"errors"`
	Output   string `json:"output"`
}

// NewSelfTester creates a new self tester
func NewSelfTester(projectDir string) *SelfTester {
	return &SelfTester{
		projectDir: projectDir,
	}
}

// RunTests runs all tests in the project
func (my *SelfTester) RunTests(ctx context.Context) (*TestResult, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "./...", "-v")
	cmd.Dir = my.projectDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()

	result := &TestResult{
		Package: "./...",
		Output:  output,
	}

	// Parse test output
	result.Tests = strings.Count(output, "=== RUN")
	result.Failures = strings.Count(output, "--- FAIL")
	result.Errors = strings.Count(output, "FAIL\t")
	result.Passed = err == nil && result.Failures == 0 && result.Errors == 0

	return result, nil
}

// RunTestsForPackage runs tests for a specific package
func (my *SelfTester) RunTestsForPackage(ctx context.Context, pkg string) (*TestResult, error) {
	cmd := exec.CommandContext(ctx, "go", "test", pkg, "-v")
	cmd.Dir = my.projectDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()

	result := &TestResult{
		Package: pkg,
		Output:  output,
	}

	result.Tests = strings.Count(output, "=== RUN")
	result.Failures = strings.Count(output, "--- FAIL")
	result.Errors = strings.Count(output, "FAIL\t")
	result.Passed = err == nil && result.Failures == 0 && result.Errors == 0

	return result, nil
}

// BuildProject builds the project to verify it compiles
func (my *SelfTester) BuildProject(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = my.projectDir
	return cmd.Run()
}

// GenerateReport generates a test report
func (my *SelfTester) GenerateReport(results []*TestResult) string {
	var sb strings.Builder
	sb.WriteString("# Self-Test Report\n\n")

	totalTests := 0
	totalFailures := 0
	for _, r := range results {
		totalTests += r.Tests
		totalFailures += r.Failures
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		sb.WriteString(fmt.Sprintf("- %s: %s (%d tests, %d failures)\n", r.Package, status, r.Tests, r.Failures))
	}

	sb.WriteString(fmt.Sprintf("\n## Summary\n"))
	sb.WriteString(fmt.Sprintf("- Total Tests: %d\n", totalTests))
	sb.WriteString(fmt.Sprintf("- Total Failures: %d\n", totalFailures))
	sb.WriteString(fmt.Sprintf("- Success Rate: %.1f%%\n", float64(totalTests-totalFailures)*100/float64(totalTests)))

	return sb.String()
}
