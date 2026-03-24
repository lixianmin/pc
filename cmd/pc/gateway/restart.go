package gateway

import (
	"fmt"
	"os"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/gateway"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the gateway daemon",
	Long: `Restart the PersonalClaw gateway daemon.

This command stops the running daemon (if any) and starts a new one.
The new daemon will load the latest compiled binary.

Examples:
  pc gateway restart`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get PC directory
		pcDir := getPCDir()

		// Create daemon manager
		daemon := gateway.NewDaemon(pcDir)

		// Check if running
		status := daemon.GetStatus()
		if status.Running {
			// Stop the daemon
			fmt.Printf("Stopping gateway (pid: %d)...\n", status.Pid)
			logo.Info("Stopping gateway daemon for restart (pid:", status.Pid, ")")

			if err := daemon.Stop(); err != nil {
				logo.Error("Failed to stop daemon:", err)
				fmt.Println("Warning: Failed to stop daemon, attempting to start anyway")
			} else {
				fmt.Println("Gateway stopped successfully")
			}

			// Wait a moment to ensure the process has fully terminated
			time.Sleep(100 * time.Millisecond)
		}

		// Start the daemon
		fmt.Println("Starting gateway...")
		logo.Info("Starting PersonalClaw gateway daemon after restart...")
		if err := daemon.Start(); err != nil {
			logo.Error("Failed to start daemon:", err)
			fmt.Println("Failed to start gateway:", err)
			os.Exit(1)
		}

		// Get new status
		status = daemon.GetStatus()
		if status.Running {
			fmt.Printf("Gateway restarted successfully (pid: %d)\n", status.Pid)
			fmt.Printf("Log file: %s\n", daemon.GetLogPath())
		} else {
			fmt.Println("Gateway may have failed to start. Check logs at:", daemon.GetLogPath())
			os.Exit(1)
		}
	},
}
