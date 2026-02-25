package logger

import (
	"os"
	"sync"

	"github.com/lixianmin/logo"
)

const (
	LevelNone  = logo.LevelNone
	LevelDebug = logo.LevelDebug
	LevelInfo  = logo.LevelInfo
	LevelWarn  = logo.LevelWarn
	LevelError = logo.LevelError
)

var (
	instance Logger
	once     sync.Once
)

// Logger defines the logging interface.
type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
	SetLevel(level string)
	Close() error
}

// Initialize creates or returns the singleton logger instance.
func Initialize(level, output string) (Logger, error) {
	var err error
	once.Do(func() {
		// Create logger instance
		loggerInstance := logo.NewLogger()
		instance = &loggerImpl{
			logger: loggerInstance,
			hooks:  make([]logo.IHook, 0),
		}

		// Parse level
		logLevel := LevelInfo
		switch level {
		case "debug":
			logLevel = LevelDebug
		case "info":
			logLevel = LevelInfo
		case "warn":
			logLevel = LevelWarn
		case "error":
			logLevel = LevelError
		}

		// Add console hook
		if output == "" || output == "stdout" || output == "-" {
			consoleHook := logo.NewConsoleHook(logo.ConsoleHookArgs{
				Flag:        logo.FlagDate | logo.FlagTime | logo.FlagLevel,
				FilterLevel: logLevel,
			})
			loggerInstance.AddHook(consoleHook)
			instance.(*loggerImpl).hooks = append(instance.(*loggerImpl).hooks, consoleHook)
		} else {
			// File output using RollingFileHook
			fileHook := logo.NewRollingFileHook(logo.RollingFileHookArgs{
				Flag:           logo.FlagDate | logo.FlagTime | logo.FlagLevel,
				FilterLevel:    logLevel,
				DirName:        output,
				FileNamePrefix: "pc",
				MaxFileSize:    10 * 1024 * 1024, // 10MB
			})
			loggerInstance.AddHook(fileHook)
			instance.(*loggerImpl).hooks = append(instance.(*loggerImpl).hooks, fileHook)

			// Also add console output for visibility
			consoleHook := logo.NewConsoleHook(logo.ConsoleHookArgs{
				Flag:        logo.FlagDate | logo.FlagTime | logo.FlagLevel,
				FilterLevel: logLevel,
			})
			loggerInstance.AddHook(consoleHook)
			instance.(*loggerImpl).hooks = append(instance.(*loggerImpl).hooks, consoleHook)
		}

		// Set as global logger
		logo.SetLogger(loggerInstance)
	})

	return instance, err
}

// Get returns the logger instance. Returns nil if not initialized.
func Get() Logger {
	return instance
}

type loggerImpl struct {
	logger *logo.Logger
	hooks  []logo.IHook
}

func (l *loggerImpl) Debug(v ...any) {
	if len(v) == 1 {
		l.logger.Debug("%v", v[0])
	} else {
		l.logger.Debug("%+v", v)
	}
}

func (l *loggerImpl) Debugf(format string, v ...any) {
	l.logger.Debug(format, v...)
}

func (l *loggerImpl) Info(v ...any) {
	if len(v) == 1 {
		l.logger.Info("%v", v[0])
	} else {
		l.logger.Info("%+v", v)
	}
}

func (l *loggerImpl) Infof(format string, v ...any) {
	l.logger.Info(format, v...)
}

func (l *loggerImpl) Warn(v ...any) {
	if len(v) == 1 {
		l.logger.Warn("%v", v[0])
	} else {
		l.logger.Warn("%+v", v)
	}
}

func (l *loggerImpl) Warnf(format string, v ...any) {
	l.logger.Warn(format, v...)
}

func (l *loggerImpl) Error(v ...any) {
	if len(v) == 1 {
		l.logger.Error("%v", v[0])
	} else {
		l.logger.Error("%+v", v)
	}
}

func (l *loggerImpl) Errorf(format string, v ...any) {
	l.logger.Error(format, v...)
}

func (l *loggerImpl) Fatal(v ...any) {
	if len(v) == 1 {
		l.logger.Debug("%v", v[0])
	} else {
		l.logger.Debug("%+v", v)
	}
	os.Exit(1)
}

func (l *loggerImpl) Fatalf(format string, v ...any) {
	l.logger.Error(format, v...)
	os.Exit(1)
}

func (l *loggerImpl) SetLevel(level string) {
	logLevel := LevelInfo
	switch level {
	case "debug":
		logLevel = LevelDebug
	case "info":
		logLevel = LevelInfo
	case "warn":
		logLevel = LevelWarn
	case "error":
		logLevel = LevelError
	}

	for _, hook := range l.hooks {
		hook.SetFilterLevel(logLevel)
	}
}

func (l *loggerImpl) Close() error {
	return l.logger.Close()
}

// AddHook adds a custom hook to the logger.
func AddHook(hook logo.IHook) {
	if instance != nil {
		if impl, ok := instance.(*loggerImpl); ok {
			impl.logger.AddHook(hook)
			impl.hooks = append(impl.hooks, hook)
		}
	}
}
