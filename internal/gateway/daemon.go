package gateway

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/config"
	"github.com/lixianmin/pc/internal/debug"
	"github.com/lixianmin/pc/internal/engine"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/pkg/types"
)

// Daemon manages the lifecycle of the gateway daemon.
type Daemon struct {
	pidfile    *Pidfile
	socketPath string
	logPath    string
}

// NewDaemon creates a new daemon manager.
func NewDaemon(pcDir string) *Daemon {
	return &Daemon{
		pidfile:    NewPidfile(filepath.Join(pcDir, "pc.pid")),
		socketPath: filepath.Join(pcDir, "pc.sock"),
		logPath:    filepath.Join(pcDir, "logs", "pc.log"),
	}
}

// GetPidfile returns the pidfile manager.
func (my *Daemon) GetPidfile() *Pidfile {
	return my.pidfile
}

// GetSocketPath returns the Unix socket path.
func (my *Daemon) GetSocketPath() string {
	return my.socketPath
}

// GetLogPath returns the log file path.
func (my *Daemon) GetLogPath() string {
	return my.logPath
}

// Status represents the daemon status.
type Status struct {
	Running   bool      `json:"running"`
	Pid       int       `json:"pid,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Start starts the daemon in the background.
func (my *Daemon) Start() error {
	// Check if already running
	if my.pidfile.IsRunning() {
		pid, _ := my.pidfile.Read()
		return fmt.Errorf("daemon is already running (pid: %d)", pid)
	}

	// Clean up stale pidfile
	if err := my.pidfile.Cleanup(); err != nil {
		return err
	}

	// Ensure log directory exists
	logDir := filepath.Dir(my.logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(my.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	// Get the executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start the daemon process
	cmd := exec.Command(exePath, "gateway", "run", "--daemon")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session, detach from terminal
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Write pidfile
	if err := my.pidfile.Write(cmd.Process.Pid); err != nil {
		// Try to kill the process if we can't write pidfile
		cmd.Process.Kill()
		return fmt.Errorf("failed to write pidfile: %w", err)
	}

	// Wait a moment and check if process is still running
	time.Sleep(100 * time.Millisecond)
	if !my.pidfile.IsRunning() {
		my.pidfile.Remove()
		return fmt.Errorf("daemon failed to start, check logs at %s", my.logPath)
	}

	return nil
}

// Stop stops the daemon.
func (my *Daemon) Stop() error {
	if !my.pidfile.Exists() {
		return fmt.Errorf("daemon is not running (no pidfile)")
	}

	pid, err := my.pidfile.Read()
	if err != nil {
		return fmt.Errorf("failed to read pidfile: %w", err)
	}

	if !my.pidfile.IsRunning() {
		// Process not running, clean up pidfile
		my.pidfile.Remove()
		return fmt.Errorf("daemon is not running (stale pidfile removed)")
	}

	// Send SIGTERM to the daemon
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait for process to exit (with timeout)
	timeout := time.AfterFunc(5*time.Second, func() {
		syscall.Kill(pid, syscall.SIGKILL)
	})
	defer timeout.Stop()

	for my.pidfile.IsRunning() {
		time.Sleep(100 * time.Millisecond)
	}

	// Remove pidfile
	my.pidfile.Remove()

	return nil
}

// GetStatus returns the current daemon status.
func (my *Daemon) GetStatus() *Status {
	status := &Status{
		Running: false,
	}

	if !my.pidfile.Exists() {
		status.Error = "pidfile not found"
		return status
	}

	pid, err := my.pidfile.Read()
	if err != nil {
		status.Error = fmt.Sprintf("failed to read pidfile: %v", err)
		return status
	}

	status.Pid = pid

	if !my.pidfile.IsRunning() {
		status.Error = "process not running (stale pidfile)"
		return status
	}

	status.Running = true
	return status
}

// Run runs the daemon process (internal use, called with --daemon flag).
func (my *Daemon) Run() error {
	pcDir := filepath.Dir(my.pidfile.path)
	pluginsDir := filepath.Join(pcDir, "plugins")

	cfg, err := config.Load(filepath.Join(pcDir, "config.yml"))
	if err != nil {
		logo.Warn("Failed to load config, using defaults:", err)
		cfg = config.DefaultConfig()
	}

	pm, err := plugin.NewPluginManager(pluginsDir)
	if err != nil {
		return fmt.Errorf("failed to create plugin manager: %w", err)
	}

	eng := engine.NewEngine(pm)

	// TEMP: 启用 BAML，阶段 3 删除此行
	eng.SetUseBAML(true)

	if cfg.GetPromptsDir() != "" {
		recorder := debug.NewPromptRecorder(cfg.GetPromptsDir(), cfg.GetMaxPromptFiles())
		eng.SetPromptRecorder(recorder)
	}

	llmPlugins := pm.GetPluginsByType(types.PluginTypeLLM)
	if len(llmPlugins) > 0 {
		eng.SetLLMPlugin(llmPlugins[0])
	}

	eng.FetchSession("default")

	server := NewRPCServer(my.socketPath, eng, pm)
	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start RPC server: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan

	server.Stop()
	eng.Close()
	my.pidfile.Remove()
	return nil
}
