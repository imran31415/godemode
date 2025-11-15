package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/imran31415/godemode/pkg/compiler"
	"github.com/imran31415/godemode/pkg/executor"
)

const (
	version = "0.1.0"
)

func main() {
	// Define CLI flags
	var (
		sourceFile    = flag.String("file", "", "Path to Go source file to execute")
		codeFlag      = flag.String("code", "", "Go source code to execute (inline)")
		timeoutFlag   = flag.Duration("timeout", 30*time.Second, "Execution timeout (e.g., 30s, 1m)")
		memoryLimitMB = flag.Uint("memory", 64, "Memory limit in MB")
		showVersion   = flag.Bool("version", false, "Show version information")
		checkTinyGo   = flag.Bool("check", false, "Check if TinyGo is installed")
		verbose       = flag.Bool("v", false, "Verbose output")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GoDeMode - Sandboxed Go Code Execution via WebAssembly\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Execute code from file\n")
		fmt.Fprintf(os.Stderr, "  %s --file examples/simple.go\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Execute code from stdin\n")
		fmt.Fprintf(os.Stderr, "  echo 'package main; import \"fmt\"; func main() { fmt.Println(\"Hello\") }' | %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Execute inline code\n")
		fmt.Fprintf(os.Stderr, "  %s --code 'package main; import \"fmt\"; func main() { fmt.Println(\"Hello\") }'\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Custom timeout and memory limit\n")
		fmt.Fprintf(os.Stderr, "  %s --file code.go --timeout 60s --memory 128\n\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("GoDeMode version %s\n", version)
		fmt.Printf("Go Code Mode: Sandboxed execution via WebAssembly\n")
		return
	}

	// Handle check flag
	if *checkTinyGo {
		if err := compiler.CheckTinyGo(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nTo install TinyGo:\n")
			fmt.Fprintf(os.Stderr, "  macOS:   brew tap tinygo-org/tools && brew install tinygo\n")
			fmt.Fprintf(os.Stderr, "  Linux:   See https://tinygo.org/getting-started/install/\n")
			os.Exit(1)
		}
		fmt.Println("✓ TinyGo is installed and accessible")
		return
	}

	// Get source code from file, flag, or stdin
	sourceCode, err := getSourceCode(*sourceFile, *codeFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading source code: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Source code length: %d bytes\n", len(sourceCode))
		fmt.Fprintf(os.Stderr, "Timeout: %v\n", *timeoutFlag)
		fmt.Fprintf(os.Stderr, "Memory limit: %d MB\n", *memoryLimitMB)
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Create executor
	exec := executor.NewExecutorWithConfig(uint32(*memoryLimitMB), *timeoutFlag)

	// Execute code
	if *verbose {
		fmt.Fprintf(os.Stderr, "Validating and compiling...\n")
	}

	ctx := context.Background()
	result, err := exec.Execute(ctx, sourceCode, *timeoutFlag)

	// Print results
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		if result != nil && result.Error != "" {
			fmt.Fprintf(os.Stderr, "Error details: %s\n", result.Error)
		}
		if result != nil && result.Stderr != "" {
			fmt.Fprintf(os.Stderr, "\nStderr:\n%s\n", result.Stderr)
		}
		os.Exit(1)
	}

	// Print stdout
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}

	// Print stderr if verbose
	if *verbose && result.Stderr != "" {
		fmt.Fprintf(os.Stderr, "\nStderr:\n%s\n", result.Stderr)
	}

	// Print execution stats if verbose
	if *verbose {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Execution time: %v\n", result.Duration)
		fmt.Fprintf(os.Stderr, "Success: %v\n", result.Success)
		fmt.Fprintf(os.Stderr, "Cache size: %d modules\n", exec.GetCacheSize())
	}
}

// getSourceCode retrieves source code from file, flag, or stdin
func getSourceCode(filePath, codeFlag string) (string, error) {
	// Priority: inline code > file > stdin

	if codeFlag != "" {
		return codeFlag, nil
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		return string(data), nil
	}

	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat stdin: %w", err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Stdin is piped
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(data), nil
	}

	// No source provided
	return "", fmt.Errorf("no source code provided (use --file, --code, or pipe to stdin)")
}
