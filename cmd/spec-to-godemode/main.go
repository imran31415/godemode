package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imran31415/godemode/pkg/codegen"
	"github.com/imran31415/godemode/pkg/spec"
)

const version = "1.0.0"

func main() {
	// Define flags
	specFile := flag.String("spec", "", "Path to MCP or OpenAPI specification file (required)")
	outputDir := flag.String("output", "./generated", "Output directory for generated code")
	packageName := flag.String("package", "tools", "Package name for generated code")
	showVersion := flag.Bool("version", false, "Show version and exit")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("spec-to-godemode version %s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if *help {
		printHelp()
		os.Exit(0)
	}

	// Validate required flags
	if *specFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -spec flag is required\n\n")
		printHelp()
		os.Exit(1)
	}

	// Run the conversion
	if err := convertSpecToGoDeMode(*specFile, *outputDir, *packageName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully generated GoDeMode code in %s\n", *outputDir)
}

func convertSpecToGoDeMode(specFile, outputDir, packageName string) error {
	// Read the spec file
	data, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	// Detect spec format
	format := spec.DetectSpecFormat(data)
	if format == spec.FormatUnknown {
		return fmt.Errorf("unknown spec format - file must be MCP or OpenAPI specification")
	}

	fmt.Printf("Detected spec format: %s\n", format)

	// Parse the spec based on format
	var tools []spec.ToolDefinition
	switch format {
	case spec.FormatMCP:
		mcpSpec, err := spec.ParseMCPSpecFromBytes(data)
		if err != nil {
			return fmt.Errorf("failed to parse MCP spec: %w", err)
		}

		// Validate the spec
		if err := mcpSpec.Validate(); err != nil {
			return fmt.Errorf("invalid MCP spec: %w", err)
		}

		tools = mcpSpec.ToToolDefinitions()
		fmt.Printf("Parsed %d tools from MCP spec '%s'\n", len(tools), mcpSpec.Name)

	case spec.FormatOpenAPI:
		openAPISpec, err := spec.ParseOpenAPISpecFromBytes(data)
		if err != nil {
			return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
		}

		tools = openAPISpec.ToToolDefinitions()
		fmt.Printf("Parsed %d tools from OpenAPI spec '%s'\n", len(tools), openAPISpec.Info.Title)

	default:
		return fmt.Errorf("unsupported spec format: %s", format)
	}

	if len(tools) == 0 {
		return fmt.Errorf("no tools found in specification")
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate code
	gen := codegen.NewCodeGenerator(packageName)

	// Generate registry.go
	fmt.Println("Generating registry.go...")
	registryCode, err := gen.GenerateRegistry(tools)
	if err != nil {
		return fmt.Errorf("failed to generate registry: %w", err)
	}

	registryPath := filepath.Join(outputDir, "registry.go")
	if err := os.WriteFile(registryPath, []byte(registryCode), 0644); err != nil {
		return fmt.Errorf("failed to write registry.go: %w", err)
	}

	// Generate tools.go
	fmt.Println("Generating tools.go...")
	toolsCode, err := gen.GenerateToolsFile(tools)
	if err != nil {
		return fmt.Errorf("failed to generate tools: %w", err)
	}

	toolsPath := filepath.Join(outputDir, "tools.go")
	if err := os.WriteFile(toolsPath, []byte(toolsCode), 0644); err != nil {
		return fmt.Errorf("failed to write tools.go: %w", err)
	}

	// Generate README.md
	fmt.Println("Generating README.md...")
	readme := gen.GenerateREADME(tools, format)
	readmePath := filepath.Join(outputDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	fmt.Printf("\nGenerated files:\n")
	fmt.Printf("  - %s\n", registryPath)
	fmt.Printf("  - %s\n", toolsPath)
	fmt.Printf("  - %s\n", readmePath)

	return nil
}

func printHelp() {
	fmt.Println("spec-to-godemode - Convert MCP or OpenAPI specs to GoDeMode tool registry")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  spec-to-godemode -spec <file> [options]")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  -spec string")
	fmt.Println("        Path to MCP or OpenAPI specification file")
	fmt.Println()
	fmt.Println("Optional Flags:")
	fmt.Println("  -output string")
	fmt.Println("        Output directory for generated code (default: ./generated)")
	fmt.Println("  -package string")
	fmt.Println("        Package name for generated code (default: tools)")
	fmt.Println("  -version")
	fmt.Println("        Show version and exit")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate from MCP spec")
	fmt.Println("  spec-to-godemode -spec mcp-server.json")
	fmt.Println()
	fmt.Println("  # Generate from OpenAPI spec with custom output")
	fmt.Println("  spec-to-godemode -spec api-spec.json -output ./mytools -package mytools")
	fmt.Println()
	fmt.Println("  # Show version")
	fmt.Println("  spec-to-godemode -version")
}
