package logger

import (
	"github.com/lixianmin/logo"
)

var (
	theLogger = logo.GetLogger().(*logo.Logger)
)

func Init(level int, logDir string) {
	if logDir == "" {
		logDir = "logs"
	}

	theLogger.SetFilterLevel(level)
	const flag = logo.FlagDate | logo.FlagTime | logo.FlagShortFile | logo.FlagLevel

	var fileHook = logo.NewRollingFileHook(
		logo.WithHookFlag(flag),
		logo.WithDirName(logDir),
		// logo.WithFileNamePrefix("pc"),
		logo.WithMaxFileSize(10*1024*1024),
	)

	theLogger.AddHook(fileHook)
}

func SetLevel(level int) {
	theLogger.SetFilterLevel(level)
}

func Get() *logo.Logger {
	return theLogger
}

func Close() error {
	if theLogger != nil {
		return theLogger.Close()
	}
	return nil
}
