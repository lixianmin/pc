package task

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Long: `Delete a task by its ID.

Examples:
  pc task delete task-001`,
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

		// Note: DeleteTask is not implemented in RPC client yet
		// For now, we'll show a message
		fmt.Printf("Delete task %s (not yet implemented)\n", id)
	},
}
