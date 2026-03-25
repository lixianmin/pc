package main

import (
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/logger"
)

func main() {
	defer loom.DumpIfPanic()
	logger.Init(logo.LevelInfo, "")

	Execute()
}
