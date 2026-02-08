package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type target struct {
	goos   string
	goarch string
	output string
}

func main() {
	targets := []target{
		{goos: "linux", goarch: "amd64", output: "writing-tools-mcp-linux-amd64"},
		{goos: "darwin", goarch: "amd64", output: "writing-tools-mcp-darwin-amd64"},
		{goos: "darwin", goarch: "arm64", output: "writing-tools-mcp-darwin-arm64"},
		{goos: "windows", goarch: "amd64", output: "writing-tools-mcp-windows-amd64.exe"},
	}

	if err := os.MkdirAll("dist", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create dist directory: %v\n", err)
		os.Exit(1)
	}

	for _, t := range targets {
		out := filepath.Join("dist", t.output)
		fmt.Printf("building %s/%s -> %s\n", t.goos, t.goarch, out)

		cmd := exec.Command("go", "build", "-o", out, ".")
		cmd.Env = append(os.Environ(),
			"CGO_ENABLED=0",
			"GOOS="+t.goos,
			"GOARCH="+t.goarch,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "build failed for %s/%s: %v\n", t.goos, t.goarch, err)
			os.Exit(1)
		}
	}

	fmt.Println("all builds completed")
}
