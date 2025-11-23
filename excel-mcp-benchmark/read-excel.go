package main

import (
	"encoding/json"
	"fmt"

	exceltools "github.com/imran31415/godemode/excel-mcp-benchmark/generated"
)

func main() {
	registry := exceltools.NewRegistry()

	// First describe the sheets
	result, _ := registry.Call("describe_sheets", map[string]interface{}{
		"file_path": "results/invoice_analysis_toolcalling.xlsx",
	})
	fmt.Println("=== Workbook Structure ===")
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(resultJSON))

	sheets := []string{"Summary", "ByRegion", "ByCategory", "BySegment", "TopCustomers"}

	for _, sheet := range sheets {
		result, err := registry.Call("read_sheet", map[string]interface{}{
			"file_path":  "results/invoice_analysis_toolcalling.xlsx",
			"sheet_name": sheet,
		})
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", sheet, err)
			continue
		}

		fmt.Printf("\n=== %s Sheet ===\n", sheet)
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(resultJSON))
	}
}
