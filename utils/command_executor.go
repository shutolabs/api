package utils

import (
	"bytes"
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
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
} 