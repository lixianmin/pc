package main

import (
	"fmt"
	"os"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/cmd/pc/gateway"
	"github.com/lixianmin/pc/cmd/pc/task"
	"github.com/lixianmin/pc/internal/wizard"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:   "pc",
		Short: "PersonalClaw - AI Agent Operating System Kernel",
		Long: `PersonalClaw (PC) is an open-source, self-hosted AI agent platform.

PC serves as the kernel for AI agents, managing their lifecycle, plugin systems,
communication protocols, and providing a foundation for autonomous digital entities.`,
		Run: func(cmd *cobra.Command, args []string) {
			showVersion, _ := cmd.Flags().GetBool("version")
			if showVersion {
				fmt.Fprintf(cmd.OutOrStdout(), "PersonalClaw version %s\n", version)
				return
			}

			configPath := wizard.GenerateConfigPath("")
			if !wizard.ConfigExists(configPath) {
				if err := wizard.RunWizard(); err != nil {
					logo.Error("Initialization failed:", err)
					os.Exit(1)
				}
				return
			}

			cmd.Help()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")
	rootCmd.AddCommand(gateway.Cmd)
	rootCmd.AddCommand(task.Cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
