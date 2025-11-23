```go
package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func main() {
	filePath := "/Users/arsheenali/dev/godemode/excel-mcp-benchmark/invoice_analysis.xlsx"

	fmt.Println("============================================================")
	fmt.Println("COMPREHENSIVE REVENUE ANALYSIS DASHBOARD")
	fmt.Println("============================================================")

	// Step 1: Read and understand the existing data
	fmt.Println("\n1. Reading existing data...")
	
	// Read Invoices sheet
	invoicesResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "Invoices",
	})
	if err != nil {
		fmt.Println("Error reading Invoices sheet:", err)
		return
	}

	// Read Customers sheet
	customersResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "Customers",
	})
	if err != nil {
		fmt.Println("Error reading Customers sheet:", err)
		return
	}

	fmt.Println("✓ Successfully read Invoices and Customers data")

	// Step 2: Create Summary sheet
	fmt.Println("\n2. Creating Summary sheet...")
	
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "Summary",
	})
	if err != nil {
		fmt.Println("Error creating Summary sheet:", err)
		return
	}

	// Write Summary headers and formulas
	summaryHeaders := [][]interface{}{
		{"Metric", "Value"},
		{"Total Revenue", "=SUM(Invoices!I:I)"},
		{"Total Invoices", "=COUNTA(Invoices!A:A)-1"},
		{"Average Invoice Value", "=AVERAGE(Invoices!I:I)"},
		{"Max Single Invoice", "=MAX(Invoices!I:I)"},
	}

	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "Summary",
		"range":      "A1:B5",
		"values":     summaryHeaders,
		"new_sheet":  false,
	})
	if err != nil {
		fmt.Println("Error writing Summary data:", err)
		return
	}

	fmt.Println("✓ Summary sheet created with key metrics")

	// Step 3: Create ByRegion sheet
	fmt.Println("\n3. Creating ByRegion analysis...")
	
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "ByRegion",
	})
	if err != nil {
		fmt.Println("Error creating ByRegion sheet:", err)
		return
	}

	regionData := [][]interface{}{
		{"Region", "TotalRevenue", "InvoiceCount", "AvgInvoice"},
		{"North", "=SUMIF(Invoices!F:F,\"North\",Invoices!I:I)", "=COUNTIF(Invoices!F:F,\"North\")", "=AVERAGEIF(Invoices!F:F,\"North\",Invoices!I:I)"},
		{"South", "=SUMIF(Invoices!F:F,\"South\",Invoices!I:I)", "=COUNTIF(Invoices!F:F,\"South\")", "=AVERAGEIF(Invoices!F:F,\"South\",Invoices!I:I)"},
		{"East", "=SUMIF(Invoices!F:F,\"East\",Invoices!I:I)", "=COUNTIF(Invoices!F:F,\"East\")", "=AVERAGEIF(Invoices!F:F,\"East\",Invoices!I:I)"},
		{"West", "=SUMIF(Invoices!F:F,\"West\",Invoices!I:I)", "=COUNTIF(Invoices!F:F,\"West\")", "=AVERAGEIF(Invoices!F:F,\"West\",Invoices!I:I)"},
	}

	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "ByRegion",
		"range":      "A1:D5",
		"values":     regionData,
		"new_sheet":  false,
	})
	if err != nil {
		fmt.Println("Error writing ByRegion data:", err)
		return
	}

	fmt.Println("✓ ByRegion analysis sheet created")

	// Step 4: Create ByCategory sheet
	fmt.Println("\n4. Creating ByCategory analysis...")
	
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "ByCategory",
	})
	if err != nil {
		fmt.Println("Error creating ByCategory sheet:", err)
		return
	}

	categoryData := [][]interface{}{
		{"Category", "TotalRevenue", "InvoiceCount"},
		{"Electronics", "=SUMIF(Invoices!E:E,\"Electronics\",Invoices!I:I)", "=COUNTIF(Invoices!E:E,\"Electronics\")"},
		{"Software", "=SUMIF(Invoices!E:E,\"Software\",Invoices!I:I)", "=COUNTIF(Invoices!E:E,\"Software\")"},
		{"Services", "=SUMIF(Invoices!E:E,\"Services\",Invoices!I:I)", "=COUNTIF(Invoices!E:E,\"Services\")"},
		{"Hardware", "=SUMIF(Invoices!E:E,\"Hardware\",Invoices!I:I)", "=COUNTIF(Invoices!E:E,\"Hardware\")"},
	}

	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "ByCategory",
		"range":      "A1:C5",
		"values":     categoryData,
		"new_sheet":  false,
	})
	if err != nil {
		fmt.Println("Error writing ByCategory data:", err)
		return
	}

	fmt.Println("✓ ByCategory analysis sheet created")

	// Step 5: Create BySegment sheet
	fmt.Println("\n5. Creating BySegment analysis...")
	
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "BySegment",
	})
	if err != nil {
		fmt.Println("Error creating BySegment sheet:", err)
		return
	}

	segmentData := [][]interface{}{
		{"Segment", "TotalRevenue", "CustomerCount"},
		{"Enterprise", "=SUMPRODUCT((VLOOKUP(Invoices!C:C,Customers!A:C,3,FALSE)=\"Enterprise\")*(Invoices!I:I))", "=COUNTIF(Customers!C:C,\"Enterprise\")"},
		{"SMB", "=SUMPRODUCT((VLOOKUP(Invoices!C:C,Customers!A:C,3,FALSE)=\"SMB\")*(Invoices!I:I))", "=COUNTIF(Customers!C:C,\"SMB\")"},
		{"Startup", "=SUMPRODUCT((VLOOKUP(Invoices!C:C,Customers!A:C,3,FALSE)=\"Startup\")*(Invoices!I:I))", "=COUNTIF(Customers!C:C,\"Startup\")"},
	}

	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "BySegment",
		"range":      "A1:C4",
		"values":     segmentData,
		"new_sheet":  false,
	})
	if err != nil {
		fmt.Println("Error writing BySegment data:", err)
		return
	}

	fmt.Println("✓ BySegment analysis sheet created")

	// Step 6: Create TopCustomers sheet
	fmt.Println("\n6. Creating TopCustomers analysis...")
	
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "TopCustomers",
	})
	if err != nil {
		fmt.Println("Error creating TopCustomers sheet:", err)
		return
	}

	topCustomersHeaders := [][]interface{}{
		{"Rank", "CustomerID", "CustomerName", "TotalSpend"},
		{"1", "=INDEX(Customers!A:A,MATCH(LARGE(SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),1),SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),0))", "=INDEX(Customers!B:B,MATCH(B2,Customers!A:A,0))", "=SUMIF(Invoices!C:C,B2,Invoices!I:I)"},
		{"2", "=INDEX(Customers!A:A,MATCH(LARGE(SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),2),SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),0))", "=INDEX(Customers!B:B,MATCH(B3,Customers!A:A,0))", "=SUMIF(Invoices!C:C,B3,Invoices!I:I)"},
		{"3", "=INDEX(Customers!A:A,MATCH(LARGE(SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),3),SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),0))", "=INDEX(Customers!B:B,MATCH(B4,Customers!A:A,0))", "=SUMIF(Invoices!C:C,B4,Invoices!I:I)"},
		{"4", "=INDEX(Customers!A:A,MATCH(LARGE(SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),4),SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),0))", "=INDEX(Customers!B:B,MATCH(B5,Customers!A:A,0))", "=SUMIF(Invoices!C:C,B5,Invoices!I:I)"},
		{"5", "=INDEX(Customers!A:A,MATCH(LARGE(SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),5),SUMIF(Invoices!C:C,Customers!A:A,Invoices!I:I),0))", "=INDEX(Customers!B:B,MATCH(B6,Customers!A:A,0))", "=SUMIF(Invoices!C:C,B6,Invoices!I:I)"},
	}

	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "TopCustomers",
		"range":      "A1:D6",
		"values":     topCustomersHeaders,
		"new_sheet":  false,
	})
	if err != nil {
		fmt.Println("Error writing TopCustomers data:", err)
		return
	}

	fmt.Println("✓ TopCustomers analysis sheet created")

	// Step 7: Create MonthlyTrend sheet
	fmt.Println("\n7. Creating MonthlyTrend analysis...")
	
	_, err = registry.Call("create_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "MonthlyTrend",
	})
	if err != nil {
		fmt.Println("Error creating MonthlyTrend sheet:", err)
		return
	}

	monthlyData := [][]interface{}{
		{"Month", "Revenue", "InvoiceCount"},
		{"Jan-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=1)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=1)*(YEAR(Invoices!B:B)=2024))"},
		{"Feb-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=2)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=2)*(YEAR(Invoices!B:B)=2024))"},
		{"Mar-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=3)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=3)*(YEAR(Invoices!B:B)=2024))"},
		{"Apr-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=4)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=4)*(YEAR(Invoices!B:B)=2024))"},
		{"May-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=5)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=5)*(YEAR(Invoices!B:B)=2024))"},
		{"Jun-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=6)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=6)*(YEAR(Invoices!B:B)=2024))"},
		{"Jul-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=7)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=7)*(YEAR(Invoices!B:B)=2024))"},
		{"Aug-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=8)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=8)*(YEAR(Invoices!B:B)=2024))"},
		{"Sep-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=9)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=9)*(YEAR(Invoices!B:B)=2024))"},
		{"Oct-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=10)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=10)*(YEAR(Invoices!B:B)=2024))"},
		{"Nov-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=11)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=11)*(YEAR(Invoices!B:B)=2024))"},
		{"Dec-2024", "=SUMPRODUCT((MONTH(Invoices!B:B)=12)*(YEAR(Invoices!B:B)=2024)*(Invoices!I:I))", "=SUMPRODUCT((MONTH(Invoices!B:B)=12)*(YEAR(Invoices!B:B)=2024))"},
	}

	_, err = registry.Call("write_to_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "MonthlyTrend",
		"range":      "A1:C13",
		"values":     monthlyData,
		"new_sheet":  false,
	})
	if err != nil {
		fmt.Println("Error writing MonthlyTrend data:", err)
		return
	}

	fmt.Println("✓ MonthlyTrend analysis sheet created")

	// Step 8: Verify calculations by reading back the created sheets
	fmt.Println("\n8. Verifying all calculations...")
	fmt.Println("============================================================")

	// Wait a moment for Excel to calculate formulas
	time.Sleep(1000)

	// Read and display Summary
	fmt.Println("\n📊 SUMMARY METRICS:")
	summaryResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "Summary",
		"range":      "A1:B5",
	})
	if err != nil {
		fmt.Println("Error reading Summary sheet:", err)
	} else {
		if data, ok := summaryResult.(map[string]interface{}); ok {
			if values, ok := data["values"].([]interface{}); ok {
				for i, row := range values {
					if rowData, ok := row.([]interface{}); ok && len(rowData) >= 2 {
						if i == 0 {
							fmt.Printf("%-25s %s\n", rowData[0], rowData[1])
							fmt.Println("----------------------------------------")
						} else {
							fmt.Printf("%-25s %v\n", rowData[0], rowData[1])
						}
					}
				}
			}
		}
	}

	// Read and display ByRegion
	fmt.Println("\n🌍 REVENUE BY REGION:")
	regionResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "ByRegion",
		"range":      "A1:D5",
	})
	if err != nil {
		fmt.Println("Error reading ByRegion sheet:", err)
	} else {
		if data, ok := regionResult.(map[string]interface{}); ok {
			if values, ok := data["values"].([]interface{}); ok {
				for i, row := range values {
					if rowData, ok := row.([]interface{}); ok && len(rowData) >= 4 {
						if i == 0 {
							fmt.Printf("%-10s %-15s %-12s %s\n", rowData[0], rowData[1], rowData[2], rowData[3])
							fmt.Println("--------------------------------------------------")
						} else {
							fmt.Printf("%-10s %-15v %-12v %v\n", rowData[0], rowData[1], rowData[2], rowData[3])
						}
					}
				}
			}
		}
	}

	// Read and display ByCategory
	fmt.Println("\n📦 REVENUE BY CATEGORY:")
	categoryResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "ByCategory",
		"range":      "A1:C5",
	})
	if err != nil {
		fmt.Println("Error reading ByCategory sheet:", err)
	} else {
		if data, ok := categoryResult.(map[string]interface{}); ok {
			if values, ok := data["values"].([]interface{}); ok {
				for i, row := range values {
					if rowData, ok := row.([]interface{}); ok && len(rowData) >= 3 {
						if i == 0 {
							fmt.Printf("%-12s %-15s %s\n", rowData[0], rowData[1], rowData[2])
							fmt.Println("---------------------------------------")
						} else {
							fmt.Printf("%-12s %-15v %v\n", rowData[0], rowData[1], rowData[2])
						}
					}
				}
			}
		}
	}

	// Read and display BySegment
	fmt.Println("\n🎯 REVENUE BY SEGMENT:")
	segmentResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "BySegment",
		"range":      "A1:C4",
	})
	if err != nil {
		fmt.Println("Error reading BySegment sheet:", err)
	} else {
		if data, ok := segmentResult.(map[string]interface{}); ok {
			if values, ok := data["values"].([]interface{}); ok {
				for i, row := range values {
					if rowData, ok := row.([]interface{}); ok && len(rowData) >= 3 {
						if i == 0 {
							fmt.Printf("%-12s %-15s %s\n", rowData[0], rowData[1], rowData[2])
							fmt.Println("---------------------------------------")
						} else {
							fmt.Printf("%-12s %-15v %v\n", rowData[0], rowData[1], rowData[2])
						}
					}
				}
			}
		}
	}

	// Read and display TopCustomers
	fmt.Println("\n🏆 TOP 5 CUSTOMERS:")
	topResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "TopCustomers",
		"range":      "A1:D6",
	})
	if err != nil {
		fmt.Println("Error reading TopCustomers sheet:", err)
	} else {
		if data, ok := topResult.(map[string]interface{}); ok {
			if values, ok := data["values"].([]interface{}); ok {
				for i, row := range values {
					if rowData, ok := row.([]interface{}); ok && len(rowData) >= 4 {
						if i == 0 {
							fmt.Printf("%-5s %-12s %-20s %s\n", rowData[0], rowData[1], rowData[2], rowData[3])
							fmt.Println("-------------------------------------------------------")
						} else {
							fmt.Printf("%-5v %-12v %-20v %v\n", rowData[0], rowData[1], rowData[2], rowData[3])
						}
					}
				}
			}
		}
	}

	// Read and display MonthlyTrend
	fmt.Println("\n📈 MONTHLY TREND (2024):")
	trendResult, err := registry.Call("read_sheet", map[string]interface{}{
		"file_path":  filePath,
		"sheet_name": "MonthlyTrend",
		"range":      "A1:C13",
	})
	if err != nil {
		fmt.Println("Error reading MonthlyTrend sheet:", err)
	} else {
		if data, ok := trendResult.(map[string]interface{}); ok {
			if values, ok := data["values"].([]interface{}); ok {
				for i, row := range values {
					if rowData, ok := row.([]interface{}); ok && len(rowData) >= 3 {
						if i == 0 {
							fmt.Printf("%-10s %-15s %s\n", rowData[0], rowData[1], rowData[2])
							fmt.Println("-------------------------------------")
						} else {
							fmt.Printf("%-10s %-15v %v\n", rowData[0], rowData[1], rowData[2])
						}
					}
				}
			}
		}
	}

	// List all sheets to confirm creation
	fmt.Println("\n📋 ALL SHEETS IN WORKBOOK:")
	sheetsResult, err := registry.Call("describe_sheets", map[string]interface{}{
		"file_path": filePath,
	})
	if err != nil {
		fmt.Println("Error describing sheets:", err)
	} else {
		if data, ok := sheetsResult.(map[string]interface{}); ok {
			if sheets, ok := data["sheets"].([]interface{}); ok {
				for i, sheet := range sheets {
					fmt.Printf("%d. %v\n", i+1, sheet)
				}
			}
		}
	}

	fmt.Println("\n============================================================")
	fmt.Println("✅ COMPREHENSIVE REVENUE ANALYSIS DASHBOARD COMPLETED!")
	fmt.Println("============================================================")
	fmt.Println("All analysis sheets have been created with dynamic formulas:")
	fmt.Println("• Summary - Key performance metrics")
	fmt.Println("• ByRegion - Regional revenue analysis")  
	fmt.Println("• ByCategory - Product category analysis")
	fmt.Println("• BySegment - Customer segment analysis")
	fmt.Println("• TopCustomers - Top 5 customers by revenue")
	fmt.Println("• MonthlyTrend - Month-by-month revenue trend")
	fmt.Println("\nAll calculations use Excel formulas for dynamic updates!")
}
```