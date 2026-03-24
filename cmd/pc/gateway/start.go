package gateway

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/gateway"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the gateway daemon",
	Long: `Start the PersonalClaw gateway daemon in the background.

The daemon will:
- Initialize the core engine and plugin system
- Load the configured agent and skills
- Start listening on Unix Socket for CLI connections
- Maintain persistent state and scheduled tasks

Examples:
  pc gateway start`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get PC directory
		pcDir := getPCDir()

		// Create daemon manager
		daemon := gateway.NewDaemon(pcDir)

		// Check if already running
		status := daemon.GetStatus()
		if status.Running {
			fmt.Printf("Gateway is already running (pid: %d)\n", status.Pid)
			os.Exit(0)
		}

		// Start the daemon
		logo.Info("Starting PersonalClaw gateway daemon...")
		if err := daemon.Start(); err != nil {
			logo.Error("Failed to start daemon:", err)
			os.Exit(1)
		}

		// Get new status
		status = daemon.GetStatus()
		if status.Running {
			fmt.Printf("Gateway started successfully (pid: %d)\n", status.Pid)
			fmt.Printf("Log file: %s\n", daemon.GetLogPath())
		} else {
			fmt.Println("Gateway may have failed to start. Check logs at:", daemon.GetLogPath())
			os.Exit(1)
		}
	},
}

// getPCDir returns the PC configuration directory.
func getPCDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".pc")
	}
	return filepath.Join(home, ".pc")
}
