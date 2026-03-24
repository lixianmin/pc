package task

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var listStatus string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List all tasks or filter by status.

Examples:
  pc task list              # List all tasks
  pc task list --status pending
  pc task list --status completed`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getRPCClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintln(os.Stderr, "Make sure the gateway is running: pc gateway start")
			os.Exit(1)
		}
		defer client.Close()

		tasks, err := client.ListTasks(listStatus)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing tasks: %v\n", err)
			os.Exit(1)
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return
		}

		// Print header
		fmt.Printf("%-12s %-10s %-30s %s\n", "ID", "Status", "Title", "Description")
		fmt.Println("--------------------------------------------------------------------------------")

		// Print tasks
		for _, task := range tasks {
			desc := task.Description
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}
			title := task.Title
			if len(title) > 30 {
				title = title[:27] + "..."
			}
			fmt.Printf("%-12s %-10s %-30s %s\n", task.ID, task.Status, title, desc)
		}
	},
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (pending, in_progress, completed)")
}
