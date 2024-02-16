package main

import (
	"log/slog"
	"os"

	"github.com/JenswBE/encrypted-paper/cmd"
	"github.com/JenswBE/encrypted-paper/utils"
)

func main() {
	// Ensure dependencies are available
	err := utils.CheckDependencies("xz", "qrencode", "zbarimg")
	if err != nil {
		slog.Error("Dependencies check failed", "error", err)
		os.Exit(1)
	}

	// Execute command
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
