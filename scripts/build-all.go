//go:build ignore

// Build script for cross-platform compilation
// Usage: go run scripts/build-all.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	version = "1.0.0"
)

var targets = []struct {
	OS   string
	Arch string
	Name string
}{
	{"windows", "amd64", "modbus-sim.exe"},
	{"linux", "amd64", "modbus-sim"},
	{"linux", "arm", "modbus-sim"},
	{"linux", "arm64", "modbus-sim"},
	{"darwin", "amd64", "modbus-sim"},
	{"darwin", "arm64", "modbus-sim"},
}

func main() {
	fmt.Println("Starting cross-platform build for Modbus-Sim...")

	// Get the project root directory
	projectRoot, err := filepath.Abs(filepath.Join("."))
	if err != nil {
		fmt.Printf("Failed to get project root: %v\n", err)
		os.Exit(1)
	}

	// Change to project root if we're in scripts directory
	if err := os.Chdir(projectRoot); err != nil {
		fmt.Printf("Failed to change directory: %v\n", err)
		os.Exit(1)
	}

	for _, target := range targets {
		outputDir := filepath.Join("build", fmt.Sprintf("%s_%s", target.OS, target.Arch))
		outputPath := filepath.Join(outputDir, target.Name)

		fmt.Printf("Building %s/%s... ", target.OS, target.Arch)

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}

		cmd := exec.Command("go", "build",
			"-ldflags", fmt.Sprintf("-X main.Version=%s", version),
			"-o", outputPath,
			".")

		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOOS=%s", target.OS),
			fmt.Sprintf("GOARCH=%s", target.Arch),
		)

		if target.OS == "linux" && target.Arch == "arm" {
			cmd.Env = append(cmd.Env, "GOARM=7")
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("FAILED: %v\n", string(output))
			continue
		}

		fmt.Println("OK")
	}

	fmt.Println("\nAll builds completed!")
	fmt.Println("Output directory: build/")
}
