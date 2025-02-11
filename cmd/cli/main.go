package main

import (
	"flag"
	"fmt"
	"os"

	"shuto-api/cmd/commands"
)

var availableCommands = map[string]commands.Command{
	"sign": commands.NewSignCommand(),
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: command is required")
		printUsage()
		os.Exit(1)
	}

	commandName := os.Args[1]
	args := os.Args[2:]

	if commandName == "-h" || commandName == "--help" {
		printUsage()
		os.Exit(0)
	}

	command, exists := availableCommands[commandName]
	if !exists {
		fmt.Printf("Unknown command: %s\n", commandName)
		printUsage()
		os.Exit(1)
	}

	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println(command.Usage())
		os.Exit(0)
	}

	command.Execute(args)
}

func printUsage() {
	fmt.Println("\nUsage:")
	fmt.Println("  shuto-cli <command> [options]")
	fmt.Println("\nAvailable Commands:")
	
	// Print commands in a consistent format
	for _, cmd := range availableCommands {
		fmt.Printf("  %-12s %s\n", cmd.Name(), cmd.Description())
	}
	
	fmt.Println("\nUse 'shuto-cli <command> -h' for help on specific commands")
}

// Reset resets the flag package state, useful for testing
func Reset() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
} 