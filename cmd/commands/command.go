package commands

// Command represents a CLI command
type Command interface {
	// Execute runs the command with the given arguments
	Execute(args []string)
	// Name returns the command name
	Name() string
	// Description returns a short description of the command
	Description() string
	// Usage returns the command usage information
	Usage() string
} 