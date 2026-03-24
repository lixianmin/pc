package gateway

import (
	"fmt"
	"os"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/gateway"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the gateway daemon",
	Long: `Stop the PersonalClaw gateway daemon gracefully.

This sends a SIGTERM signal to the daemon process and waits for it to shut down.
If the daemon doesn't stop within 5 seconds, it will be forcefully killed.

Examples:
  pc gateway stop`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get PC directory
		pcDir := getPCDir()

		// Create daemon manager
		daemon := gateway.NewDaemon(pcDir)

		// Check if running
		status := daemon.GetStatus()
		if !status.Running {
			fmt.Println("Gateway is not running")
			if status.Error != "" {
				fmt.Printf("Note: %s\n", status.Error)
			}
			os.Exit(0)
		}

		// Stop the daemon
		fmt.Printf("Stopping gateway (pid: %d)...\n", status.Pid)
		logo.Info("Stopping gateway daemon (pid:", status.Pid, ")")

		if err := daemon.Stop(); err != nil {
			logo.Error("Failed to stop daemon:", err)
			os.Exit(1)
		}

		fmt.Println("Gateway stopped successfully")
	},
}
