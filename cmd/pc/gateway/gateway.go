package gateway

import (
	"github.com/spf13/cobra"
)

// Cmd is the gateway command
var Cmd = &cobra.Command{
	Use:   "gateway",
	Short: "Manage the PersonalClaw gateway daemon",
	Long: `Manage the PersonalClaw gateway daemon which runs in the background.

The gateway daemon is responsible for:
- Managing the AI agent lifecycle
- Loading and coordinating plugins
- Handling RPC requests from CLI and other clients
- Maintaining persistent state and task scheduling

Examples:
  pc gateway start    # Start the gateway daemon
  pc gateway stop     # Stop the gateway daemon
  pc gateway status   # Check daemon status`,
}

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(restartCmd)
	Cmd.AddCommand(statusCmd)
}
