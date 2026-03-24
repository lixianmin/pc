package task

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete <id>",
	Short: "Mark a task as completed",
	Long: `Mark a task as completed by its ID.

Examples:
  pc task complete task-001`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		client, err := getRPCClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintln(os.Stderr, "Make sure the gateway is running: pc gateway start")
			os.Exit(1)
		}
		defer client.Close()

		if err := client.CompleteTask(id); err != nil {
			fmt.Fprintf(os.Stderr, "Error completing task: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Task %s marked as completed\n", id)
	},
}
