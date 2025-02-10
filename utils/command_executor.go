package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Mock for CommandExecutor
type MockCommandExecutor struct {
	ExecuteFunc func(command string, args ...string) ([]byte, error)
}

func (m *MockCommandExecutor) Execute(command string, args ...string) ([]byte, error) {
	return m.ExecuteFunc(command, args...)
}

// CommandExecutor defines an interface for executing commands
type CommandExecutor interface {
	Execute(command string, args ...string) ([]byte, error)
}

// execCommand is the default implementation of CommandExecutor
type execCommand struct{}

func NewCommandExecutor() CommandExecutor {
	return &execCommand{}
}

// Execute runs the command and returns the output
func (e *execCommand) Execute(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("command failed: %w: %s", err, stderr.String())
		}
		return nil, fmt.Errorf("command failed: %w", err)
	}

	if stderr.Len() > 0 {
		Debug("Command produced stderr output", "command", command, "stderr", stderr.String())
	}

	return stdout.Bytes(), nil
} 