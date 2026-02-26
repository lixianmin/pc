package main

import (
	"fmt"
	"os"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/logger"
	"github.com/lixianmin/pc/internal/wizard"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("PersonalClaw version %s\n", version)
		return
	}

	// Initialize logger with default settings
	logger.Init("info")

	// Check if config exists, run wizard if not
	configPath := wizard.GenerateConfigPath("")
	if !wizard.ConfigExists(configPath) {
		if err := wizard.RunWizard(); err != nil {
			logo.Error("Initialization failed:", err)
			os.Exit(1)
		}
		return
	}

	// Config exists, start normally
	logo.Info("PersonalClaw - AI Agent Operating System Kernel")
	logo.Info("Version: %s\n", version)
	logo.Info("Configuration loaded from:", configPath)
}
