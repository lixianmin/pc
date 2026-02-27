package logger

import (
	"strings"

	"github.com/lixianmin/logo"
)

var (
	theLogger = logo.GetLogger().(*logo.Logger)
)

// Init initializes the logger with given level and output.
func Init(level string) {
	// Set filter level
	filterLevel := getFilterLevel(level)
	theLogger.SetFilterLevel(filterLevel)

	// var consoleHook = logo.NewConsoleHook(logo.ConsoleHookArgs{Flag: logo.FlagDate | logo.FlagTime | logo.FlagLevel})
	// theLogger.AddHook(consoleHook)

	var fileHook = logo.NewRollingFileHook(logo.RollingFileHookArgs{
		Flag:           logo.FlagDate | logo.FlagTime | logo.FlagLevel,
		DirName:        "logs",
		FileNamePrefix: "pc",
		MaxFileSize:    10 * 1024 * 1024, // 10MB
	})

	theLogger.AddHook(fileHook)
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
	return theLogger
}

// Close closes the logger.
func Close() error {
	if theLogger != nil {
		return theLogger.Close()
	}
	return nil
}
