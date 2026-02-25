package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("PersonalClaw version %s\n", version)
		return
	}

	fmt.Println("PersonalClaw - AI Agent Operating System Kernel")
	fmt.Printf("Version: %s\n", version)
}
