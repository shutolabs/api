package utils

import (
	"bytes"
	"testing"
)

// Test for MockCommandExecutor
func TestMockCommandExecutor(t *testing.T) {
	mock := &MockCommandExecutor{
		ExecuteFunc: func(command string, args ...string) ([]byte, error) {
			return []byte("mock output"), nil
		},
	}

	output, err := mock.Execute("mockCommand", "arg1", "arg2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(output) != "mock output" {
		t.Fatalf("expected 'mock output', got %s", output)
	}
}

// Test for execCommand
func TestExecCommand(t *testing.T) {
	e := &execCommand{}

	// This test will fail if the command is not available on the system
	output, err := e.Execute("echo", "Hello, World!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !bytes.Contains(output, []byte("Hello, World!")) {
		t.Fatalf("expected output to contain 'Hello, World!', got %s", output)
	}
}

// Test for execCommand with an invalid command
func TestExecCommandInvalid(t *testing.T) {
	e := &execCommand{}

	_, err := e.Execute("invalidCommand")
	if err == nil {
		t.Fatal("expected an error for invalid command, got none")
	}
} 