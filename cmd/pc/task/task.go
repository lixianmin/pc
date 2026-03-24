package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lixianmin/pc/internal/gateway"
	"github.com/spf13/cobra"
)

// Cmd is the task command
var Cmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long: `Manage tasks in PersonalClaw.

Tasks are tracked in task_list.md and can be created, listed, completed, and deleted.

Examples:
  pc task list              # List all tasks
  pc task list --status pending  # List pending tasks
  pc task add "My task"     # Add a new task
  pc task complete <id>     # Mark a task as completed
  pc task delete <id>       # Delete a task`,
}

// getRPCClient creates an RPC client connected to the gateway
func getRPCClient() (*gateway.RpcClient, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	pcDir := filepath.Join(home, ".pc")

	daemon := gateway.NewDaemon(pcDir)
	status := daemon.GetStatus()
	if !status.Running {
		return nil, fmt.Errorf("gateway is not running")
	}

	client := gateway.NewRpcClient(daemon.GetSocketPath())
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	return client, nil
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(completeCmd)
	Cmd.AddCommand(deleteCmd)
}
