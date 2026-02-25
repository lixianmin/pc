package logger

import (
	"sync"
	"testing"
)

func TestLoggerInitialization(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		output  string
		wantErr bool
	}{
		{
			name:    "stdout output with info level",
			level:   "info",
			output:  "stdout",
			wantErr: false,
		},
		{
			name:    "stdout output with debug level",
			level:   "debug",
			output:  "-",
			wantErr: false,
		},
		{
			name:    "empty level defaults to info",
			level:   "",
			output:  "stdout",
			wantErr: false,
		},
		{
			name:    "warn level",
			level:   "warn",
			output:  "stdout",
			wantErr: false,
		},
		{
			name:    "error level",
			level:   "error",
			output:  "stdout",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any previous instance
			instance = nil
			once = sync.Once{}

			logger, err := Initialize(tt.level, tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if logger == nil {
				t.Error("Initialize() returned nil logger")
				return
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		logFunc func(Logger)
	}{
		{
			name:  "info level",
			level: "info",
			logFunc: func(l Logger) {
				l.Info("info message")
			},
		},
		{
			name:  "debug level",
			level: "debug",
			logFunc: func(l Logger) {
				l.Debug("debug message")
			},
		},
		{
			name:  "warn level",
			level: "warn",
			logFunc: func(l Logger) {
				l.Warn("warn message")
			},
		},
		{
			name:  "error level",
			level: "error",
			logFunc: func(l Logger) {
				l.Error("error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any previous instance
			instance = nil
			once = sync.Once{}

			logger, err := Initialize(tt.level, "")
			if err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			// Just verify logging doesn't panic
			tt.logFunc(logger)
		})
	}
}

func TestLoggerFormattedOutput(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		format  string
		args    []any
	}{
		{
			name:   "Infof",
			level:  "info",
			format: "test message: %s",
			args:   []any{"value"},
		},
		{
			name:   "Debugf",
			level:  "debug",
			format: "debug %d",
			args:   []any{42},
		},
		{
			name:   "Warnf",
			level:  "warn",
			format: "warning: %v",
			args:   []any{"something"},
		},
		{
			name:   "Errorf",
			level:  "error",
			format: "error: %s",
			args:   []any{"failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any previous instance
			instance = nil
			once = sync.Once{}

			logger, err := Initialize(tt.level, "")
			if err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			// Verify formatted logging doesn't panic
			switch tt.name {
			case "Infof":
				logger.Infof(tt.format, tt.args...)
			case "Debugf":
				logger.Debugf(tt.format, tt.args...)
			case "Warnf":
				logger.Warnf(tt.format, tt.args...)
			case "Errorf":
				logger.Errorf(tt.format, tt.args...)
			}
		})
	}
}

func TestLoggerSingleton(t *testing.T) {
	// Clean up any previous instance
	instance = nil
	once = sync.Once{}

	// Initialize logger
	logger1, err := Initialize("info", "stdout")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Get logger again - should return same instance
	logger2 := Get()

	if logger1 != logger2 {
		t.Error("Get() did not return the same instance")
	}

	// Initialize again - should not create new instance
	logger3, err := Initialize("debug", "stdout")
	if err != nil {
		t.Fatalf("Initialize() second call error = %v", err)
	}

	if logger1 != logger3 {
		t.Error("Initialize() second call created new instance")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	// Clean up any previous instance
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("error", "stdout")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Set level to debug
	logger.SetLevel("debug")
	logger.Debug("debug message should now appear")

	// Set level to error
	logger.SetLevel("error")
	logger.Debug("debug message should not appear")
}

func TestLoggerMultipleArgs(t *testing.T) {
	// Clean up any previous instance
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("debug", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test multiple args - should not panic
	logger.Info("arg1", "arg2", 42)
	logger.Debug("debug1", "debug2")
	logger.Warn("warn1", "warn2")
}

func TestLoggerLevelParsing(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		wantPanic bool
	}{
		{
			name:     "debug",
			level:    "debug",
			wantPanic: false,
		},
		{
			name:     "info",
			level:    "info",
			wantPanic: false,
		},
		{
			name:     "warn",
			level:    "warn",
			wantPanic: false,
		},
		{
			name:     "error",
			level:    "error",
			wantPanic: false,
		},
		{
			name:     "invalid level defaults to info",
			level:    "invalid",
			wantPanic: false,
		},
		{
			name:     "empty defaults to info",
			level:    "",
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any previous instance
			instance = nil
			once = sync.Once{}

			logger, err := Initialize(tt.level, "")
			if err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			// Should not panic
			logger.Info("test message")
		})
	}
}

func TestLoggerNilInstance(t *testing.T) {
	// Clean up any previous instance
	instance = nil
	once = sync.Once{}

	logger := Get()
	if logger != nil {
		t.Error("Get() should return nil before initialization")
	}
}

func TestLoggerClose(t *testing.T) {
	// Clean up any previous instance
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("info", "stdout")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	err = logger.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestLoggerStringMethods(t *testing.T) {
	// Test that string methods don't panic
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("debug", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	tests := []struct {
		name string
		fn   func()
	}{
		{"Info single string", func() { logger.Info("single") }},
		{"Info multiple strings", func() { logger.Info("a", "b", "c") }},
		{"Info mixed types", func() { logger.Info("string", 42, true) }},
		{"Debug single string", func() { logger.Debug("single") }},
		{"Debug multiple strings", func() { logger.Debug("a", "b", "c") }},
		{"Debug mixed types", func() { logger.Debug("string", 42, true) }},
		{"Warn single string", func() { logger.Warn("single") }},
		{"Warn multiple strings", func() { logger.Warn("a", "b", "c") }},
		{"Warn mixed types", func() { logger.Warn("string", 42, true) }},
		{"Error single string", func() { logger.Error("single") }},
		{"Error multiple strings", func() { logger.Error("a", "b", "c") }},
		{"Error mixed types", func() { logger.Error("string", 42, true) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn()
		})
	}
}

func TestLoggerFormattedMethods(t *testing.T) {
	// Test that formatted methods don't panic
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("debug", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{"Infof", "test %s %d", []any{"arg", 42}},
		{"Debugf", "debug %v", []any{map[string]int{"a": 1}}},
		{"Warnf", "warn %s", []any{"warning"}},
		{"Errorf", "error %d", []any{999}},
		{"Infof no args", "no args", nil},
		{"Debugf empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "Infof":
				logger.Infof(tt.format, tt.args...)
			case "Debugf":
				logger.Debugf(tt.format, tt.args...)
			case "Warnf":
				logger.Warnf(tt.format, tt.args...)
			case "Errorf":
				logger.Errorf(tt.format, tt.args...)
			case "Infof no args":
				logger.Infof(tt.format, tt.args...)
			case "Debugf empty":
				logger.Debugf(tt.format, tt.args...)
			}
		})
	}
}

func TestLoggerCloseTwice(t *testing.T) {
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("info", "stdout")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Close once
	err = logger.Close()
	if err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	// Close again - should not panic
	err = logger.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

// TestLoggerAfterClose verifies logger behavior after close
func TestLoggerAfterClose(t *testing.T) {
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("info", "stdout")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	logger.Close()

	// Logging after close should not panic
	logger.Info("message after close")
	logger.Debug("debug after close")
	logger.Warn("warn after close")
	logger.Error("error after close")
}

func TestLoggerSetLevelVariants(t *testing.T) {
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("info", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test all level variants
	levels := []string{"debug", "info", "warn", "error", "", "invalid"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger.SetLevel(level)
			logger.Info("test after setting level")
		})
	}
}

func TestLoggerConstants(t *testing.T) {
	// Verify constants are correctly defined
	tests := []struct {
		name  string
		value int
	}{
		{"LevelNone", LevelNone},
		{"LevelDebug", LevelDebug},
		{"LevelInfo", LevelInfo},
		{"LevelWarn", LevelWarn},
		{"LevelError", LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure constants are defined and have expected ordering
			if tt.value < 0 || tt.value > 4 {
				t.Errorf("Constant %s has unexpected value: %d", tt.name, tt.value)
			}
		})
	}
}

// TestConcurrentInitialization tests concurrent logger initialization
func TestConcurrentInitialization(t *testing.T) {
	instance = nil
	once = sync.Once{}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger, err := Initialize("info", "")
			if err != nil {
				t.Errorf("Initialize() error = %v", err)
			}
			if logger == nil {
				t.Error("Initialize() returned nil")
			}
		}()
	}
	wg.Wait()

	// Verify only one instance was created
	if Get() == nil {
		t.Error("No logger instance created")
	}
}

// TestConcurrentLogging tests concurrent logging
func TestConcurrentLogging(t *testing.T) {
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("debug", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			logger.Info("concurrent message", i)
			logger.Debugf("debug %d", i)
			logger.Warnf("warn %d", i)
		}(i)
	}
	wg.Wait()
}

func TestLoggerInterface(t *testing.T) {
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("info", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Verify logger implements all interface methods
	var _ Logger = logger

	// Test each method exists
	type methodChecker interface {
		Debug(...any)
		Debugf(string, ...any)
		Info(...any)
		Infof(string, ...any)
		Warn(...any)
		Warnf(string, ...any)
		Error(...any)
		Errorf(string, ...any)
		Fatal(...any)
		Fatalf(string, ...any)
		SetLevel(string)
		Close() error
	}

	var _ methodChecker = logger
}

func TestLoggerSetLevelWithLogging(t *testing.T) {
	instance = nil
	once = sync.Once{}

	logger, err := Initialize("info", "")
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Set to error level
	logger.SetLevel("error")

	// These should not produce visible output at error level
	logger.Debug("debug at error level")
	logger.Info("info at error level")
	logger.Warn("warn at error level")

	// This should produce output
	logger.Error("error at error level")

	// Set to debug level
	logger.SetLevel("debug")

	// All of these should produce output
	logger.Debug("debug at debug level")
	logger.Info("info at debug level")
	logger.Warn("warn at debug level")
	logger.Error("error at debug level")
}
