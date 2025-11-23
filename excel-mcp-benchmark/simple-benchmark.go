package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imran31415/godemode/pkg/executor"
	exceltools "github.com/imran31415/godemode/excel-mcp-benchmark/generated"
)

// API types
type ClaudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ContentBlock struct {
	Type      string      `json:"type"`
	Text      string      `json:"text,omitempty"`
	ID        string      `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	Input     interface{} `json:"input,omitempty"`
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   string      `json:"content,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

type ClaudeResponse struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type AuditEntry struct {
	Timestamp    time.Time
	Type         string
	ToolName     string
	ToolArgs     string
	ToolResult   string
	Details      string
	InputTokens  int
	OutputTokens int
}

type BenchmarkResult struct {
	Approach      string
	Duration      time.Duration
	APICallCount  int
	InputTokens   int
	OutputTokens  int
	TotalTokens   int
	ToolCalls     int
	Success       bool
	Error         string
	EstimatedCost float64
	AuditLog      []AuditEntry
	GeneratedCode string
}

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("ANTHROPIC_API_KEY not set")
		os.Exit(1)
	}

	fmt.Println("======================================================================")
	fmt.Println("Excel MCP Benchmark: CodeMode vs Native Tool Calling")
	fmt.Println("======================================================================")

	// Create test Excel file
	testFile, err := setupTestData()
	if err != nil {
		fmt.Printf("Failed to setup test data: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(testFile)

	// Simple Scenario
	fmt.Println("\n======================================================================")
	fmt.Println("SCENARIO: Simple")
	fmt.Println("======================================================================")
	fmt.Println("Expected operations: ~5 tool calls")

	simplePrompt := fmt.Sprintf(`You have access to an Excel file at: %s

Please perform the following tasks:
1. List all sheets in the workbook
2. Read the data from the "Sales" sheet
3. Get the value of cell B2 (should be a sales amount)
4. Set a formula in cell D6 to calculate the total of column B (=SUM(B2:B5))
5. Read back the Sales sheet to verify the formula was added

Output results at each step.`, testFile)

	fmt.Println("\nRunning CodeMode approach...")
	codeModeResult := runCodeModeBenchmark(apiKey, simplePrompt)
	printResult(codeModeResult)

	// Reset test file for fair comparison
	os.Remove(testFile)
	testFile, _ = setupTestData()
	defer os.Remove(testFile)

	fmt.Println("\nRunning Native Tool Calling approach...")
	toolCallingResult := runToolCallingBenchmark(apiKey, simplePrompt)
	printResult(toolCallingResult)

	printComparison(codeModeResult, toolCallingResult)

	// Complex Scenario
	fmt.Println("\n======================================================================")
	fmt.Println("SCENARIO: Complex")
	fmt.Println("======================================================================")
	fmt.Println("Expected operations: ~12 tool calls")

	// Reset test file
	os.Remove(testFile)
	testFile, _ = setupTestData()
	defer os.Remove(testFile)

	complexPrompt := fmt.Sprintf(`You have access to an Excel file at: %s

Please perform a comprehensive analysis and modification:
1. List all sheets to understand the workbook structure
2. Read the "Sales" sheet to see the sales data
3. Read the "Products" sheet to see product information
4. Create a new sheet called "Analysis"
5. In the Analysis sheet, write headers: "Metric", "Value" in A1:B1
6. Calculate and write the following metrics:
   - Total Sales (sum of Sales!B2:B5) in row 2
   - Average Sale in row 3
   - Max Sale in row 4
   - Number of Products (count from Products sheet) in row 5
7. Read back the Analysis sheet to verify
8. Copy the Sales sheet to a new sheet called "Sales_Backup"
9. Describe all sheets to show the final structure

Output detailed results at each step.`, testFile)

	fmt.Println("\nRunning CodeMode approach...")
	codeModeComplex := runCodeModeBenchmark(apiKey, complexPrompt)
	printResult(codeModeComplex)

	// Reset test file
	os.Remove(testFile)
	testFile, _ = setupTestData()

	fmt.Println("\nRunning Native Tool Calling approach...")
	toolCallingComplex := runToolCallingBenchmark(apiKey, complexPrompt)
	printResult(toolCallingComplex)

	printComparison(codeModeComplex, toolCallingComplex)

	// Advanced Invoice Analysis Scenario
	fmt.Println("\n======================================================================")
	fmt.Println("SCENARIO: Advanced Invoice Analysis")
	fmt.Println("======================================================================")
	fmt.Println("Expected operations: ~30-40 tool calls")

	// Create invoice dataset
	os.Remove(testFile)
	invoiceFile, expectedValues := setupInvoiceData()
	cwd, _ := os.Getwd()

	invoicePrompt := fmt.Sprintf(`You have access to an Excel file at: %s

This file contains company invoice data with the following sheets:
- "Invoices": 100 invoices with columns: InvoiceID, Date, CustomerID, Product, Category, Region, Quantity, UnitPrice, Total
- "Customers": Customer details with columns: CustomerID, CustomerName, Segment (Enterprise/SMB/Startup)

Your task is to create a comprehensive revenue analysis dashboard. Perform these steps:

1. First, read and understand the Invoices and Customers data
2. Create a "Summary" sheet with overall metrics:
   - Total Revenue (sum of all invoice totals)
   - Total Invoices (count)
   - Average Invoice Value
   - Max Single Invoice
3. Create a "ByRegion" sheet with revenue breakdown by region:
   - Columns: Region, TotalRevenue, InvoiceCount, AvgInvoice
   - Regions: North, South, East, West
4. Create a "ByCategory" sheet with revenue by product category:
   - Columns: Category, TotalRevenue, InvoiceCount
   - Categories: Electronics, Software, Services, Hardware
5. Create a "BySegment" sheet with revenue by customer segment:
   - Columns: Segment, TotalRevenue, CustomerCount
   - Segments: Enterprise, SMB, Startup
6. Create a "TopCustomers" sheet listing top 5 customers by total spend
7. Create a "MonthlyTrend" sheet with monthly revenue (Jan-Dec 2024)
8. Verify all calculations by reading back the created sheets

Use formulas where possible (SUMIF, COUNTIF, AVERAGEIF) for dynamic calculations.
Output detailed results showing the values in each analysis sheet.`, invoiceFile)

	fmt.Println("\nRunning CodeMode approach...")
	codeModeInvoice := runCodeModeBenchmark(apiKey, invoicePrompt)
	printResult(codeModeInvoice)

	// Validate results if successful
	if codeModeInvoice.Success {
		validateInvoiceResults(invoiceFile, expectedValues, "CodeMode")
	}

	// Save CodeMode result for review
	codeModeSavedFile := filepath.Join(cwd, "results", "invoice_analysis_codemode.xlsx")
	os.MkdirAll(filepath.Join(cwd, "results"), 0755)
	copyFile(invoiceFile, codeModeSavedFile)
	fmt.Printf("    📁 Saved CodeMode output: %s\n", codeModeSavedFile)

	// Reset for tool calling
	os.Remove(invoiceFile)
	invoiceFile, expectedValues = setupInvoiceData()

	fmt.Println("\nRunning Native Tool Calling approach...")
	toolCallingInvoice := runToolCallingBenchmark(apiKey, invoicePrompt)
	printResult(toolCallingInvoice)

	// Validate results
	if toolCallingInvoice.Success {
		validateInvoiceResults(invoiceFile, expectedValues, "Tool Calling")
	}

	// Save Tool Calling result for review
	toolCallingSavedFile := filepath.Join(cwd, "results", "invoice_analysis_toolcalling.xlsx")
	copyFile(invoiceFile, toolCallingSavedFile)
	fmt.Printf("    📁 Saved Tool Calling output: %s\n", toolCallingSavedFile)

	// Clean up
	os.Remove(invoiceFile)

	printComparison(codeModeInvoice, toolCallingInvoice)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// ExpectedValues holds the expected calculation results for validation
type ExpectedValues struct {
	TotalRevenue    float64
	TotalInvoices   int
	AvgInvoice      float64
	MaxInvoice      float64
	RegionRevenue   map[string]float64
	CategoryRevenue map[string]float64
	SegmentRevenue  map[string]float64
}

func setupInvoiceData() (string, ExpectedValues) {
	registry := exceltools.NewRegistry()

	cwd, _ := os.Getwd()
	invoiceFile := filepath.Join(cwd, "invoice_analysis.xlsx")

	// Create workbook
	registry.Call("create_workbook", map[string]interface{}{
		"file_path": invoiceFile,
	})

	// Generate 100 realistic invoices
	invoices := generateInvoiceData()
	customers := generateCustomerData()

	// Write Invoices sheet
	invoiceData := []interface{}{
		[]interface{}{"InvoiceID", "Date", "CustomerID", "Product", "Category", "Region", "Quantity", "UnitPrice", "Total"},
	}
	for _, inv := range invoices {
		invoiceData = append(invoiceData, []interface{}{
			inv.ID, inv.Date, inv.CustomerID, inv.Product, inv.Category, inv.Region, inv.Quantity, inv.UnitPrice, inv.Total,
		})
	}

	registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  invoiceFile,
		"sheet_name": "Sheet1",
		"range":      fmt.Sprintf("A1:I%d", len(invoiceData)),
		"values":     invoiceData,
	})

	// Rename to Invoices
	registry.Call("create_sheet", map[string]interface{}{
		"file_path":  invoiceFile,
		"sheet_name": "Invoices",
	})
	registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  invoiceFile,
		"sheet_name": "Invoices",
		"range":      fmt.Sprintf("A1:I%d", len(invoiceData)),
		"values":     invoiceData,
	})

	// Write Customers sheet
	customerData := []interface{}{
		[]interface{}{"CustomerID", "CustomerName", "Segment"},
	}
	for _, cust := range customers {
		customerData = append(customerData, []interface{}{
			cust.ID, cust.Name, cust.Segment,
		})
	}

	registry.Call("create_sheet", map[string]interface{}{
		"file_path":  invoiceFile,
		"sheet_name": "Customers",
	})
	registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  invoiceFile,
		"sheet_name": "Customers",
		"range":      fmt.Sprintf("A1:C%d", len(customerData)),
		"values":     customerData,
	})

	// Calculate expected values for validation
	expected := calculateExpectedValues(invoices, customers)

	return invoiceFile, expected
}

type Invoice struct {
	ID         string
	Date       string
	CustomerID string
	Product    string
	Category   string
	Region     string
	Quantity   int
	UnitPrice  float64
	Total      float64
}

type Customer struct {
	ID      string
	Name    string
	Segment string
}

func generateInvoiceData() []Invoice {
	products := []struct {
		name     string
		category string
		price    float64
	}{
		{"Laptop Pro", "Electronics", 1299.99},
		{"Desktop PC", "Hardware", 899.99},
		{"Cloud Suite", "Software", 499.99},
		{"Consulting", "Services", 150.00},
		{"Monitor 27", "Electronics", 349.99},
		{"Server Rack", "Hardware", 2499.99},
		{"Security SW", "Software", 299.99},
		{"Training", "Services", 200.00},
		{"Tablet X", "Electronics", 599.99},
		{"NAS Storage", "Hardware", 799.99},
	}

	regions := []string{"North", "South", "East", "West"}
	months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}

	var invoices []Invoice

	for i := 1; i <= 100; i++ {
		prod := products[i%len(products)]
		region := regions[i%len(regions)]
		month := months[i%len(months)]
		day := fmt.Sprintf("%02d", (i%28)+1)
		qty := (i % 5) + 1
		custID := fmt.Sprintf("CUST%03d", (i%20)+1)

		inv := Invoice{
			ID:         fmt.Sprintf("INV%04d", i),
			Date:       fmt.Sprintf("2024-%s-%s", month, day),
			CustomerID: custID,
			Product:    prod.name,
			Category:   prod.category,
			Region:     region,
			Quantity:   qty,
			UnitPrice:  prod.price,
			Total:      float64(qty) * prod.price,
		}
		invoices = append(invoices, inv)
	}

	return invoices
}

func generateCustomerData() []Customer {
	segments := []string{"Enterprise", "SMB", "Startup"}
	var customers []Customer

	for i := 1; i <= 20; i++ {
		cust := Customer{
			ID:      fmt.Sprintf("CUST%03d", i),
			Name:    fmt.Sprintf("Customer %d Inc", i),
			Segment: segments[i%len(segments)],
		}
		customers = append(customers, cust)
	}

	return customers
}

func calculateExpectedValues(invoices []Invoice, customers []Customer) ExpectedValues {
	expected := ExpectedValues{
		RegionRevenue:   make(map[string]float64),
		CategoryRevenue: make(map[string]float64),
		SegmentRevenue:  make(map[string]float64),
	}

	// Create customer segment lookup
	customerSegment := make(map[string]string)
	for _, c := range customers {
		customerSegment[c.ID] = c.Segment
	}

	for _, inv := range invoices {
		expected.TotalRevenue += inv.Total
		expected.TotalInvoices++
		if inv.Total > expected.MaxInvoice {
			expected.MaxInvoice = inv.Total
		}

		expected.RegionRevenue[inv.Region] += inv.Total
		expected.CategoryRevenue[inv.Category] += inv.Total

		segment := customerSegment[inv.CustomerID]
		expected.SegmentRevenue[segment] += inv.Total
	}

	expected.AvgInvoice = expected.TotalRevenue / float64(expected.TotalInvoices)

	return expected
}

func validateInvoiceResults(invoiceFile string, expected ExpectedValues, approach string) {
	registry := exceltools.NewRegistry()

	fmt.Printf("\n    --- Validation Results (%s) ---\n", approach)

	// Read Summary sheet
	result, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  invoiceFile,
		"sheet_name": "Summary",
	})

	if err != nil {
		fmt.Printf("    ❌ Could not read Summary sheet: %v\n", err)
	} else {
		fmt.Printf("    ✅ Summary sheet created\n")
		fmt.Printf("       Expected Total Revenue: $%.2f\n", expected.TotalRevenue)
		fmt.Printf("       Expected Total Invoices: %d\n", expected.TotalInvoices)
		fmt.Printf("       Expected Avg Invoice: $%.2f\n", expected.AvgInvoice)
		fmt.Printf("       Expected Max Invoice: $%.2f\n", expected.MaxInvoice)

		// Print actual data if available
		if data, ok := result.(map[string]interface{}); ok {
			if rows, ok := data["data"].([]interface{}); ok {
				fmt.Printf("       Actual data rows: %d\n", len(rows))
			}
		}
	}

	// Check other sheets exist
	sheets := []string{"ByRegion", "ByCategory", "BySegment", "TopCustomers", "MonthlyTrend"}
	for _, sheet := range sheets {
		_, err := registry.Call("read_sheet", map[string]interface{}{
			"file_path":  invoiceFile,
			"sheet_name": sheet,
		})
		if err != nil {
			fmt.Printf("    ❌ %s sheet: not created\n", sheet)
		} else {
			fmt.Printf("    ✅ %s sheet: created\n", sheet)
		}
	}

	fmt.Println()
}

func setupTestData() (string, error) {
	registry := exceltools.NewRegistry()

	// Get absolute path
	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "test_workbook.xlsx")

	// Create workbook
	_, err := registry.Call("create_workbook", map[string]interface{}{
		"file_path": testFile,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create workbook: %w", err)
	}

	// Add Sales data
	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  testFile,
		"sheet_name": "Sheet1",
		"range":      "A1:B5",
		"values": []interface{}{
			[]interface{}{"Product", "Amount"},
			[]interface{}{"Widget A", 150.50},
			[]interface{}{"Widget B", 275.00},
			[]interface{}{"Gadget X", 99.99},
			[]interface{}{"Gadget Y", 450.00},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to write sales data: %w", err)
	}

	// Rename Sheet1 to Sales by creating new sheet and copying
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  testFile,
		"sheet_name": "Sales",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Sales sheet: %w", err)
	}

	// Write to Sales sheet directly
	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  testFile,
		"sheet_name": "Sales",
		"range":      "A1:B5",
		"values": []interface{}{
			[]interface{}{"Product", "Amount"},
			[]interface{}{"Widget A", 150.50},
			[]interface{}{"Widget B", 275.00},
			[]interface{}{"Gadget X", 99.99},
			[]interface{}{"Gadget Y", 450.00},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to write to Sales sheet: %w", err)
	}

	// Create Products sheet
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  testFile,
		"sheet_name": "Products",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Products sheet: %w", err)
	}

	// Add product details
	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  testFile,
		"sheet_name": "Products",
		"range":      "A1:C4",
		"values": []interface{}{
			[]interface{}{"Name", "Category", "Stock"},
			[]interface{}{"Widget A", "Widgets", 100},
			[]interface{}{"Widget B", "Widgets", 50},
			[]interface{}{"Gadget X", "Gadgets", 25},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to write products data: %w", err)
	}

	return testFile, nil
}

func runCodeModeBenchmark(apiKey string, prompt string) BenchmarkResult {
	start := time.Now()
	registry := exceltools.NewRegistry()
	var auditLog []AuditEntry
	var executedToolCalls int

	// Build system prompt with available tools for CodeMode
	systemPrompt := buildCodeModeSystemPrompt(registry)

	fullPrompt := systemPrompt + "\n\nTask:\n" + prompt

	// Single API call to generate code
	auditLog = append(auditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "api_call",
		Details:   "Sending prompt to Claude for code generation",
	})

	resp, err := callClaude(apiKey, fullPrompt, nil)
	if err != nil {
		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "error",
			Details:   fmt.Sprintf("API call failed: %s", err.Error()),
		})
		return BenchmarkResult{
			Approach: "CodeMode",
			Duration: time.Since(start),
			Success:  false,
			Error:    err.Error(),
			AuditLog: auditLog,
		}
	}

	auditLog = append(auditLog, AuditEntry{
		Timestamp:    time.Now(),
		Type:         "api_response",
		Details:      "Received code generation response",
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
	})

	// Extract generated code
	generatedCode := ""
	for _, block := range resp.Content {
		if block.Type == "text" {
			generatedCode = block.Text
		}
	}

	// Extract just the Go code from markdown
	goCode := extractGoCode(generatedCode)

	// Count registry.Call occurrences in generated code
	toolCalls := strings.Count(goCode, "registry.Call")
	if toolCalls == 0 {
		toolCalls = strings.Count(goCode, `Call("`)
	}

	// Log the tool calls found in generated code
	auditLog = append(auditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "code_analysis",
		Details:   fmt.Sprintf("Generated code contains %d tool calls", toolCalls),
	})

	// Execute the generated code
	auditLog = append(auditLog, AuditEntry{
		Timestamp: time.Now(),
		Type:      "execution",
		Details:   "Starting code execution via Yaegi interpreter",
	})

	output, execToolCalls, execErr := executeGeneratedCode(goCode, registry, &auditLog)
	executedToolCalls = execToolCalls

	if execErr != nil {
		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "error",
			Details:   fmt.Sprintf("Execution failed: %s", execErr.Error()),
		})
	} else {
		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "execution_complete",
			Details:   fmt.Sprintf("Execution completed with %d tool calls", executedToolCalls),
		})
		if len(output) > 500 {
			output = output[:500] + "..."
		}
		if output != "" {
			auditLog = append(auditLog, AuditEntry{
				Timestamp:  time.Now(),
				Type:       "output",
				ToolResult: output,
			})
		}
	}

	// Calculate cost (Claude Sonnet pricing: $3/MTok in, $15/MTok out)
	inputCost := float64(resp.Usage.InputTokens) * 0.003 / 1000
	outputCost := float64(resp.Usage.OutputTokens) * 0.015 / 1000

	return BenchmarkResult{
		Approach:      "CodeMode",
		Duration:      time.Since(start),
		APICallCount:  1,
		InputTokens:   resp.Usage.InputTokens,
		OutputTokens:  resp.Usage.OutputTokens,
		TotalTokens:   resp.Usage.InputTokens + resp.Usage.OutputTokens,
		ToolCalls:     executedToolCalls,
		Success:       execErr == nil,
		Error:         func() string { if execErr != nil { return execErr.Error() }; return "" }(),
		EstimatedCost: inputCost + outputCost,
		AuditLog:      auditLog,
		GeneratedCode: generatedCode,
	}
}

// extractGoCode uses the core preprocessor to extract Go code from markdown
func extractGoCode(text string) string {
	preprocessor := executor.NewCodePreprocessor()
	return preprocessor.ExtractGoCode(text)
}

// executeGeneratedCode runs the generated Go code using the InterpreterExecutor
func executeGeneratedCode(code string, registry *exceltools.Registry, auditLog *[]AuditEntry) (string, int, error) {
	// Track tool calls
	toolCallCount := 0

	// Create a wrapper registry that logs calls
	wrappedCall := func(name string, args map[string]interface{}) (interface{}, error) {
		toolCallCount++
		argsJSON, _ := json.Marshal(args)

		*auditLog = append(*auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "tool_call",
			ToolName:  name,
			ToolArgs:  string(argsJSON),
			Details:   fmt.Sprintf("Executed tool call #%d", toolCallCount),
		})

		result, err := registry.Call(name, args)

		var resultStr string
		if err != nil {
			resultStr = fmt.Sprintf("Error: %v", err)
		} else {
			resultBytes, _ := json.Marshal(result)
			resultStr = string(resultBytes)
			if len(resultStr) > 200 {
				resultStr = resultStr[:200] + "..."
			}
		}

		*auditLog = append(*auditLog, AuditEntry{
			Timestamp:  time.Now(),
			Type:       "tool_result",
			ToolName:   name,
			ToolResult: resultStr,
		})

		return result, err
	}

	// Use the core executor's ExecuteGeneratedCode API
	exec := executor.NewInterpreterExecutor()
	result, err := exec.ExecuteGeneratedCode(context.Background(), code, 60*time.Second, wrappedCall)

	if err != nil {
		return result.Stdout, toolCallCount, fmt.Errorf("execution error: %w", err)
	}

	if !result.Success {
		return result.Stdout, toolCallCount, fmt.Errorf("execution failed: %s", result.Error)
	}

	return result.Stdout, toolCallCount, nil
}

func runToolCallingBenchmark(apiKey string, prompt string) BenchmarkResult {
	start := time.Now()
	registry := exceltools.NewRegistry()
	var auditLog []AuditEntry

	tools := buildTools(registry)
	messages := []Message{
		{Role: "user", Content: prompt},
	}

	totalInputTokens := 0
	totalOutputTokens := 0
	apiCallCount := 0
	toolCallCount := 0

	// Iterative tool calling loop
	for {
		apiCallCount++

		auditLog = append(auditLog, AuditEntry{
			Timestamp: time.Now(),
			Type:      "api_call",
			Details:   fmt.Sprintf("API call #%d", apiCallCount),
		})

		resp, err := callClaudeWithTools(apiKey, messages, tools)
		if err != nil {
			return BenchmarkResult{
				Approach:     "ToolCalling",
				Duration:     time.Since(start),
				APICallCount: apiCallCount,
				Success:      false,
				Error:        err.Error(),
				AuditLog:     auditLog,
			}
		}

		totalInputTokens += resp.Usage.InputTokens
		totalOutputTokens += resp.Usage.OutputTokens

		auditLog = append(auditLog, AuditEntry{
			Timestamp:    time.Now(),
			Type:         "api_response",
			Details:      fmt.Sprintf("Response #%d (stop_reason: %s)", apiCallCount, resp.StopReason),
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		})

		// Process tool uses
		var toolResults []ContentBlock
		hasToolUse := false

		for _, block := range resp.Content {
			if block.Type == "tool_use" {
				hasToolUse = true
				toolCallCount++

				args, _ := block.Input.(map[string]interface{})
				argsJSON, _ := json.Marshal(args)

				auditLog = append(auditLog, AuditEntry{
					Timestamp: time.Now(),
					Type:      "tool_call",
					ToolName:  block.Name,
					ToolArgs:  string(argsJSON),
					Details:   fmt.Sprintf("Tool call #%d", toolCallCount),
				})

				result, err := registry.Call(block.Name, args)
				var resultStr string
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultBytes, _ := json.Marshal(result)
					resultStr = string(resultBytes)
				}

				truncatedResult := resultStr
				if len(truncatedResult) > 200 {
					truncatedResult = truncatedResult[:200] + "..."
				}

				auditLog = append(auditLog, AuditEntry{
					Timestamp:  time.Now(),
					Type:       "tool_result",
					ToolName:   block.Name,
					ToolResult: truncatedResult,
				})

				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   resultStr,
				})
			}
		}

		if !hasToolUse {
			break
		}

		// Add assistant response and tool results
		messages = append(messages, Message{
			Role:    "assistant",
			Content: resp.Content,
		})
		messages = append(messages, Message{
			Role:    "user",
			Content: toolResults,
		})

		if apiCallCount > 20 {
			break
		}
	}

	inputCost := float64(totalInputTokens) * 0.003 / 1000
	outputCost := float64(totalOutputTokens) * 0.015 / 1000

	return BenchmarkResult{
		Approach:      "ToolCalling",
		Duration:      time.Since(start),
		APICallCount:  apiCallCount,
		InputTokens:   totalInputTokens,
		OutputTokens:  totalOutputTokens,
		TotalTokens:   totalInputTokens + totalOutputTokens,
		ToolCalls:     toolCallCount,
		Success:       true,
		EstimatedCost: inputCost + outputCost,
		AuditLog:      auditLog,
	}
}

func buildCodeModeSystemPrompt(registry *exceltools.Registry) string {
	var sb strings.Builder

	sb.WriteString(`You are a code generation assistant. Generate complete, executable Go code to accomplish the user's task.

The code will be executed in an environment with access to Excel file operations through a tool registry.

Available tools (call via registry.Call("tool_name", args)):

`)

	for _, tool := range registry.ListTools() {
		sb.WriteString(fmt.Sprintf("## %s\n%s\n", tool.Name, tool.Description))
		if len(tool.Parameters) > 0 {
			sb.WriteString("Parameters:\n")
			for _, p := range tool.Parameters {
				req := ""
				if p.Required {
					req = " (required)"
				}
				sb.WriteString(fmt.Sprintf("  - %s: %s%s\n", p.Name, p.Type, req))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`
Generate a complete, valid Go program that:
1. Uses the registry to call the necessary tools
2. Implements loops for iterating over results
3. Handles errors appropriately
4. Outputs results using fmt.Println

IMPORTANT:
- Use valid Go syntax only. Do NOT use Python-style string multiplication.
- If you use strings.Repeat(), you MUST import "strings" in your import block.
- Better yet, just use fmt.Println("============================================================") for separators.
- The registry variable is already defined - do NOT redefine it.

The registry variable is already defined - do NOT redefine it. Just use registry.Call() directly.

Example usage:
  result, err := registry.Call("describe_sheets", map[string]interface{}{
      "file_path": "/path/to/file.xlsx",
  })
  if err != nil {
      fmt.Println("Error:", err)
      return
  }

  // Type assert to access the data
  if data, ok := result.(map[string]interface{}); ok {
      if sheets, ok := data["sheets"].([]interface{}); ok {
          for _, sheet := range sheets {
              fmt.Println(sheet)
          }
      }
  }
`)

	return sb.String()
}

func buildTools(registry *exceltools.Registry) []Tool {
	var tools []Tool

	for _, tool := range registry.ListTools() {
		properties := make(map[string]interface{})
		required := []string{}

		for _, p := range tool.Parameters {
			propType := "string"
			if p.Type == "boolean" {
				propType = "boolean"
			} else if p.Type == "array" {
				propType = "array"
			} else if p.Type == "any" {
				propType = "string"
			}

			properties[p.Name] = map[string]string{
				"type":        propType,
				"description": p.Description,
			}

			if p.Required {
				required = append(required, p.Name)
			}
		}

		tools = append(tools, Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: InputSchema{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		})
	}

	return tools
}

func callClaude(apiKey string, prompt string, tools []Tool) (*ClaudeResponse, error) {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	req := ClaudeRequest{
		Model:     model,
		MaxTokens: 8192,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	if tools != nil {
		req.Tools = tools
	}

	return sendRequest(apiKey, req)
}

func callClaudeWithTools(apiKey string, messages []Message, tools []Tool) (*ClaudeResponse, error) {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	req := ClaudeRequest{
		Model:     model,
		MaxTokens: 8192,
		Messages:  messages,
		Tools:     tools,
	}

	return sendRequest(apiKey, req)
}

func sendRequest(apiKey string, req ClaudeRequest) (*ClaudeResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", httpResp.StatusCode, string(body))
	}

	var resp ClaudeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func printResult(r BenchmarkResult) {
	status := "✅"
	if !r.Success {
		status = "❌"
	}

	fmt.Printf("\n  %s %s:\n", status, r.Approach)
	fmt.Printf("    Duration:     %v\n", r.Duration.Round(time.Millisecond))
	fmt.Printf("    API Calls:    %d\n", r.APICallCount)
	fmt.Printf("    Tool Calls:   %d\n", r.ToolCalls)
	fmt.Printf("    Tokens:       %d (in: %d, out: %d)\n", r.TotalTokens, r.InputTokens, r.OutputTokens)
	fmt.Printf("    Est. Cost:    $%.4f\n", r.EstimatedCost)
	if r.Error != "" {
		fmt.Printf("    Error:        %s\n", r.Error)
	}

	// Print audit log
	if len(r.AuditLog) > 0 {
		fmt.Printf("\n    --- Audit Log ---\n")
		for i, entry := range r.AuditLog {
			timestamp := entry.Timestamp.Format("15:04:05.000")
			switch entry.Type {
			case "api_call":
				fmt.Printf("    [%s] %d. API_CALL: %s\n", timestamp, i+1, entry.Details)
			case "api_response":
				fmt.Printf("    [%s] %d. API_RESPONSE: %s (tokens: in=%d, out=%d)\n",
					timestamp, i+1, entry.Details, entry.InputTokens, entry.OutputTokens)
			case "tool_call":
				fmt.Printf("    [%s] %d. TOOL_CALL: %s\n", timestamp, i+1, entry.Details)
				fmt.Printf("                    Tool: %s\n", entry.ToolName)
				args := entry.ToolArgs
				if len(args) > 100 {
					args = args[:100] + "..."
				}
				fmt.Printf("                    Args: %s\n", args)
			case "tool_result":
				fmt.Printf("    [%s] %d. TOOL_RESULT: %s\n", timestamp, i+1, entry.ToolName)
				fmt.Printf("                    Result: %s\n", entry.ToolResult)
			case "code_analysis":
				fmt.Printf("    [%s] %d. CODE_ANALYSIS: %s\n", timestamp, i+1, entry.Details)
			case "execution":
				fmt.Printf("    [%s] %d. EXECUTION: %s\n", timestamp, i+1, entry.Details)
			case "execution_complete":
				fmt.Printf("    [%s] %d. EXECUTION_COMPLETE: %s\n", timestamp, i+1, entry.Details)
			case "output":
				fmt.Printf("    [%s] %d. OUTPUT:\n%s\n", timestamp, i+1, entry.ToolResult)
			case "error":
				fmt.Printf("    [%s] %d. ERROR: %s\n", timestamp, i+1, entry.Details)
			default:
				fmt.Printf("    [%s] %d. %s: %s\n", timestamp, i+1, entry.Type, entry.Details)
			}
		}
		fmt.Println()
	}

	// Print generated code for CodeMode (truncated) and save full version
	if r.GeneratedCode != "" {
		fmt.Printf("    --- Generated Code (truncated) ---\n")
		code := r.GeneratedCode
		lines := strings.Split(code, "\n")
		maxLines := 30
		if len(lines) > maxLines {
			for i := 0; i < maxLines; i++ {
				line := lines[i]
				if len(line) > 100 {
					line = line[:100] + "..."
				}
				fmt.Printf("    %s\n", line)
			}
			fmt.Printf("    ... (%d more lines)\n", len(lines)-maxLines)
		} else {
			for _, line := range lines {
				if len(line) > 100 {
					line = line[:100] + "..."
				}
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println()

		// Save full generated code to file
		filename := fmt.Sprintf("generated-code-%s.go", strings.ToLower(strings.ReplaceAll(r.Approach, " ", "-")))
		os.WriteFile(filename, []byte(r.GeneratedCode), 0644)
		fmt.Printf("    Full code saved to: %s\n\n", filename)
	}
}

func printComparison(codeMode, toolCalling BenchmarkResult) {
	fmt.Println("\n======================================================================")
	fmt.Println("COMPARISON")
	fmt.Println("======================================================================")

	if codeMode.Success && toolCalling.Success {
		speedup := float64(toolCalling.Duration) / float64(codeMode.Duration)
		tokenReduction := 100 * (1 - float64(codeMode.TotalTokens)/float64(toolCalling.TotalTokens))
		costSavings := 100 * (1 - codeMode.EstimatedCost/toolCalling.EstimatedCost)

		fmt.Printf("\nCodeMode vs Tool Calling:\n")
		fmt.Printf("  Speed:         %.2fx faster\n", speedup)
		fmt.Printf("  Tokens:        %.1f%% fewer tokens\n", tokenReduction)
		fmt.Printf("  Cost:          %.1f%% cheaper\n", costSavings)
		fmt.Printf("  API Calls:     %d vs %d\n", codeMode.APICallCount, toolCalling.APICallCount)
	}
}
