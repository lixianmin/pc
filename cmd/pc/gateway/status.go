package gateway

import (
	"fmt"
	"os"

	"github.com/lixianmin/pc/internal/gateway"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check gateway daemon status",
	Long: `Check the current status of the PersonalClaw gateway daemon.

This command reads the pidfile and checks if the daemon process is running.

Examples:
  pc gateway status`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get PC directory
		pcDir := getPCDir()

		// Create daemon manager
		daemon := gateway.NewDaemon(pcDir)

		// Get status
		status := daemon.GetStatus()

		// Output status
		if status.Running {
			fmt.Printf("Status: running\n")
			fmt.Printf("PID: %d\n", status.Pid)
			fmt.Printf("Log: %s\n", daemon.GetLogPath())
			fmt.Printf("Socket: %s\n", daemon.GetSocketPath())
		} else {
			fmt.Printf("Status: not running\n")
			if status.Error != "" {
				fmt.Printf("Error: %s\n", status.Error)
			}
			os.Exit(1)
		}
	},
}
