// Modbus-Sim - Modbus RTU/TCP Data Simulation CLI Tool
//
// This is the main entry point for the application.
// It initializes i18n and executes the CLI commands.
package main

import (
	"modbus-sim/cmd"
	"modbus-sim/internal/i18n"
)

// Version and commit variables can be set at build time:
//
//	go build -ldflags "-X main.Version=1.0.0 -X main.Commit=abc123" -o modbus-sim .
var (
	Version = "dev"
	Commit  = "none"
)

func main() {
	// Initialize i18n with English as default language
	i18n.MustInit("en")

	// Set version variables for CLI commands
	cmd.Version = Version
	cmd.Commit = Commit

	// Execute CLI
	cmd.Execute()
}