package logger

import (
	"strings"

	"github.com/lixianmin/logo"
)

var (
	_logger = logo.GetLogger().(*logo.Logger)

	// Predefined hooks
	consoleHook = logo.NewConsoleHook(logo.ConsoleHookArgs{Flag: logo.FlagDate | logo.FlagTime | logo.FlagLevel})
	fileHook    = logo.NewRollingFileHook(logo.RollingFileHookArgs{
		Flag:           logo.FlagDate | logo.FlagTime | logo.FlagLevel,
		DirName:        "logs",
		FileNamePrefix: "pc",
		MaxFileSize:    10 * 1024 * 1024, // 10MB
	})
)

// Init initializes the logger with given level and output.
func Init(level, output string) {
	// Set filter level
	filterLevel := getFilterLevel(level)
	_logger.SetFilterLevel(filterLevel)

	// Set output based on configuration
	if output == "" || output == "stdout" || output == "-" {
		// Console output
		_logger.AddHook(consoleHook)
	} else {
		// File output
		_logger.AddHook(fileHook)
	}
}

// getFilterLevel converts string level to logo level.
func getFilterLevel(filterLevel string) int {
	level := strings.ToLower(filterLevel)
	switch level {
	case "debug":
		return logo.LevelDebug
	case "warn", "warning":
		return logo.LevelWarn
	case "error":
		return logo.LevelError
	default:
		return logo.LevelInfo
	}
}

// Get returns the logger instance.
func Get() *logo.Logger {
	return _logger
}

// Close closes the logger.
func Close() error {
	if _logger != nil {
		return _logger.Close()
	}
	return nil
}
