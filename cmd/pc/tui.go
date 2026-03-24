package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/gateway"
	"github.com/lixianmin/pc/internal/tui"
	"github.com/spf13/cobra"
)

var tuiTimeout int

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the TUI (Terminal User Interface)",
	Long: `Start the interactive Terminal User Interface for PersonalClaw.

The TUI provides a chat-like interface to interact with your AI agent:
- Multi-turn conversations with context
- Command history (up/down arrows)
- Tab completion for commands
- Slash commands (/quit, /clear, /skills, /status, /help)

Examples:
  pc tui           # Start the TUI
  pc tui --timeout 60  # Use 60 second timeout for LLM responses`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get PC directory
		home, err := os.UserHomeDir()
		if err != nil {
			logo.Error("Failed to get home directory:", err)
			os.Exit(1)
		}
		pcDir := filepath.Join(home, ".pc")

		// Check if gateway is running
		daemon := gateway.NewDaemon(pcDir)
		status := daemon.GetStatus()

		if !status.Running {
			logo.Error("Gateway is not running. Please start it first:")
			fmt.Fprintln(os.Stderr, "Gateway is not running. Please start it first:")
			fmt.Fprintln(os.Stderr, "  pc gateway start")
			os.Exit(1)
		}

		// Connect to RPC server
		client := gateway.NewRpcClient(daemon.GetSocketPath())
		client.SetTimeout(time.Duration(tuiTimeout) * time.Second)
		if err := client.Connect(); err != nil {
			logo.Error("Failed to connect to gateway:", err)
			fmt.Fprintln(os.Stderr, "Failed to connect to gateway:", err)
			fmt.Fprintln(os.Stderr, "Try restarting the gateway:")
			fmt.Fprintln(os.Stderr, "  pc gateway stop")
			fmt.Fprintln(os.Stderr, "  pc gateway start")
			os.Exit(1)
		}
		defer client.Close()

		// Create and run TUI
		model := tui.NewModel(client)
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			logo.Error("Failed to start TUI:", err)
			fmt.Fprintln(os.Stderr, "Failed to start TUI:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().IntVar(&tuiTimeout, "timeout", 60, "Timeout in seconds for LLM responses")
}
