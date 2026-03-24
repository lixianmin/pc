package gateway

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/agent"
	"github.com/lixianmin/pc/internal/config"
	"github.com/lixianmin/pc/internal/debug"
	"github.com/lixianmin/pc/internal/engine"
	"github.com/lixianmin/pc/internal/gateway"
	"github.com/lixianmin/pc/internal/logger"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/internal/wizard"
	"github.com/spf13/cobra"
)

var (
	daemonFlag bool
)

var runCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run the gateway (internal use)",
	Long:   `Run the PersonalClaw gateway. This command is used internally by the daemon.`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if !daemonFlag {
			fmt.Println("Error: This command is for internal use only. Use 'pc gateway start' instead.")
			os.Exit(1)
		}

		var pcDir = getPCDir()
		var logDir = filepath.Join(pcDir, "logs")

		logger.Init(logo.LevelInfo, logDir)

		configPath := wizard.GenerateConfigPath("")
		if !wizard.ConfigExists(configPath) {
			logo.Error("Config file not found:", configPath)
			logo.Error("Please run 'pc' first to initialize configuration")
			os.Exit(1)
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			logo.Error("Failed to load config:", err)
			os.Exit(1)
		}

		logger.SetLevel(cfg.GetLogoLevel())

		logo.Info("Starting PersonalClaw Gateway Daemon")
		logo.Info("Config:", configPath)

		agentManager, err := agent.NewManager(cfg, configPath)
		if err != nil {
			logo.Error("Failed to create agent manager:", err)
			os.Exit(1)
		}

		if err := agentManager.LoadAgentsMd(cfg.GetAgentPath()); err != nil {
			logo.Error("Failed to load agent.md:", err)
		} else {
			logo.Info("Agent configuration loaded from:", cfg.GetAgentPath())
		}

		if err := agentManager.LoadState(); err != nil {
			logo.Error("Failed to load agent state:", err)
		} else {
			logo.Info("Agent state loaded, name:", agentManager.GetAgent().Name)
		}

		pluginDir := wizard.GeneratePluginDir("")
		pluginManager, err := plugin.NewPluginManager(pluginDir)
		if err != nil {
			logo.Error("Failed to create plugin manager:", err)
			os.Exit(1)
		}

		pluginManager.SetLLMTimeout(time.Duration(cfg.GetLLMTimeout()) * time.Second)
		logo.Info("LLM timeout set to", cfg.GetLLMTimeout(), "seconds")

		if _, err := pluginManager.Discover(); err != nil {
			logo.Error("Failed to discover plugins:", err)
			os.Exit(1)
		}

		plugins := pluginManager.ListPlugins()
		logo.Info("Discovered", len(plugins), "plugins")

		daemon := gateway.NewDaemon(pcDir)
		llmPlugins := pluginManager.GetPluginsByType("llm")

		var engine = engine.NewEngine(pluginManager)

		promptsDir := cfg.GetPromptsDir()
		if promptsDir != "" {
			recorder := debug.NewPromptRecorder(promptsDir, cfg.GetMaxPromptFiles())
			engine.SetPromptRecorder(recorder)
			logo.Info("Prompt recorder enabled, dir:", promptsDir)
		}

		dynamicPrompt := engine.BuildSystemPrompt()
		logo.Info("System prompt configured (length:", len(dynamicPrompt), ")")

		skillsDir := cfg.GetSkillsDir()
		if skillsDir != "" {
			if err := engine.SetSkillDir(skillsDir); err != nil {
				logo.Warn("Failed to load skills:", err)
			} else {
				skills := engine.ListSkills()
				logo.Info("Loaded", len(skills), "skills from", skillsDir)
			}
		}

		if err := daemon.GetPidfile().Write(os.Getpid()); err != nil {
			logo.Error("Failed to write pidfile:", err)
			os.Exit(1)
		}

		logo.Info("Gateway daemon started (pid:", os.Getpid(), ")")

		var rpcServer = gateway.NewRPCServer(daemon.GetSocketPath(), engine, pluginManager)
		if err := rpcServer.Start(); err != nil {
			logo.Error("Failed to start RPC server:", err)
			os.Exit(1)
		}

		logo.Info("RPC server started on:", daemon.GetSocketPath())

		for _, p := range plugins {
			if p.Type == "llm" && p.Enabled {
				engine.SetLLMPlugin(p)
				logo.Info("Using LLM plugin:", p.Name)

				if err := pluginManager.StartPlugin(p); err != nil {
					logo.Error("Failed to pre-start LLM plugin:", err)
				} else {
					logo.Info("LLM plugin pre-started successfully")
				}
				break
			}
		}

		if len(llmPlugins) == 0 {
			logo.Warn("No LLM plugin found, echoing messages")
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

		sig := <-sigChan
		logo.Info("Received signal:", sig)

		logo.Info("Shutting down gateway daemon...")
		rpcServer.Stop()
		engine.Close()

		if err := agentManager.SaveState(); err != nil {
			logo.Error("Failed to save agent state:", err)
		} else {
			logo.Info("Agent state saved")
		}

		daemon.GetPidfile().Remove()

		logo.Info("Gateway daemon stopped")
	},
}

func init() {
	runCmd.Flags().BoolVar(&daemonFlag, "daemon", false, "Run as daemon (internal use)")
	Cmd.AddCommand(runCmd)
}
