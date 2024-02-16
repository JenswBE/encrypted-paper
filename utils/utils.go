package utils

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
)

func checkDependency(cmd string) error {
	command := exec.Command("which", cmd)
	if err := command.Run(); err != nil {
		return fmt.Errorf("required dependency is not installed: %s", cmd)
	}
	return nil
}

func CheckDependencies(cmds ...string) error {
	// Check dependencies
	missing := []string{}
	for _, cmd := range cmds {
		err := checkDependency(cmd)
		if err != nil {
			missing = append(missing, cmd)
		}
	}

	// Exit if dependency is missing
	if len(missing) > 0 {
		return fmt.Errorf("required dependencies missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

func RunCommand(description string, input io.Reader, output io.Writer, cmd string, cmdArgs ...string) error {
	command := exec.Command(cmd, cmdArgs...)
	if input != nil {
		command.Stdin = input
	}
	if output != nil {
		command.Stdout = output
	}
	var errBuff bytes.Buffer
	command.Stderr = &errBuff
	if err := command.Run(); err != nil {
		slog.Error("Command failed", "command", command.String(), "stderr", strings.TrimSpace(errBuff.String()))
		return fmt.Errorf("failed to %s: %w", description, err)
	}
	return nil
}
