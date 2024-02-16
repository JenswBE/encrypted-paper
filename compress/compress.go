package compress

import (
	"fmt"
	"io"
	"os/exec"
)

func Compress(input io.Reader, output io.Writer) error {
	return runCommandWithStdinStdout("compress", input, output, "xz", "--compress", "-9", "--extreme", "--stdout")
}

func Decompress(input io.Reader, output io.Writer) error {
	return runCommandWithStdinStdout("decompress", input, output, "xz", "--decompress", "--stdout")
}

func runCommandWithStdinStdout(description string, input io.Reader, output io.Writer, cmd string, cmdArgs ...string) error {
	command := exec.Command(cmd, cmdArgs...)
	command.Stdin = input
	command.Stdout = output
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to %s: %w", description, err)
	}
	return nil
}
