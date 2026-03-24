package task

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var taskDescription string

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new task",
	Long: `Add a new task to the task list.

Examples:
  pc task add "Review code"
  pc task add "Deploy to production" --description "Deploy v1.2.0"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]

		client, err := getRPCClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintln(os.Stderr, "Make sure the gateway is running: pc gateway start")
			os.Exit(1)
		}
		defer client.Close()

		id, err := client.AddTask(title, taskDescription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding task: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Task created: %s\n", id)
	},
}

func init() {
	addCmd.Flags().StringVar(&taskDescription, "description", "", "Task description")
}
