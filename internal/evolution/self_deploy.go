package evolution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// SelfDeployer handles self-deployment and restart
type SelfDeployer struct {
	projectDir string
}

// NewSelfDeployer creates a new self deployer
func NewSelfDeployer(projectDir string) *SelfDeployer {
	return &SelfDeployer{
		projectDir: projectDir,
	}
}

// Build builds the project
func (my *SelfDeployer) Build(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "pc", "./cmd/pc")
	cmd.Dir = my.projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %w\n%s", err, string(output))
	}
	return nil
}

// Restart restarts the application
func (my *SelfDeployer) Restart(ctx context.Context) error {
	// This is a placeholder - actual restart logic depends on deployment method
	// For now, just verify the binary exists
	binaryPath := my.projectDir + "/pc"
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}
	return nil
}

// Deploy performs full deployment: build, test, and restart
func (my *SelfDeployer) Deploy(ctx context.Context) error {
	// Build
	if err := my.Build(ctx); err != nil {
		return fmt.Errorf("deploy build failed: %w", err)
	}

	// Run tests
	tester := NewSelfTester(my.projectDir)
	result, err := tester.RunTests(ctx)
	if err != nil {
		return fmt.Errorf("deploy tests failed: %w", err)
	}

	if !result.Passed {
		return fmt.Errorf("tests failed, aborting deployment")
	}

	// Restart
	if err := my.Restart(ctx); err != nil {
		return fmt.Errorf("deploy restart failed: %w", err)
	}

	return nil
}

// GracefulShutdown performs graceful shutdown
func (my *SelfDeployer) GracefulShutdown() error {
	// Placeholder for graceful shutdown logic
	return nil
}
