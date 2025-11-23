package exceltools

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

func getString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(args map[string]interface{}, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// describeSheets lists all sheet information
func describeSheets(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	if filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	result := make([]map[string]interface{}, 0, len(sheets))

	for i, name := range sheets {
		sheetInfo := map[string]interface{}{
			"index": i,
			"name":  name,
		}

		// Get dimensions if possible
		cols, err := f.GetCols(name)
		if err == nil {
			sheetInfo["columns"] = len(cols)
		}

		rows, err := f.GetRows(name)
		if err == nil {
			sheetInfo["rows"] = len(rows)
		}

		result = append(result, sheetInfo)
	}

	return map[string]interface{}{
		"sheets": result,
		"count":  len(sheets),
	}, nil
}

// readSheet reads values from a sheet
func readSheet(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	sheetName := getString(args, "sheet_name")
	cellRange := getString(args, "range")

	if filePath == "" || sheetName == "" {
		return nil, fmt.Errorf("file_path and sheet_name are required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var data [][]string

	if cellRange != "" {
		// Parse range and read specific cells
		parts := strings.Split(cellRange, ":")
		if len(parts) == 2 {
			startCol, startRow, _ := excelize.CellNameToCoordinates(parts[0])
			endCol, endRow, _ := excelize.CellNameToCoordinates(parts[1])

			for row := startRow; row <= endRow; row++ {
				rowData := make([]string, 0)
				for col := startCol; col <= endCol; col++ {
					cellName, _ := excelize.CoordinatesToCellName(col, row)
					value, _ := f.GetCellValue(sheetName, cellName)
					rowData = append(rowData, value)
				}
				data = append(data, rowData)
			}
		}
	} else {
		// Read all rows
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to read sheet: %w", err)
		}
		data = rows
	}

	return map[string]interface{}{
		"data":  data,
		"rows":  len(data),
		"range": cellRange,
	}, nil
}

// writeToSheet writes values to a sheet
func writeToSheet(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	sheetName := getString(args, "sheet_name")
	cellRange := getString(args, "range")
	newSheet := getBool(args, "new_sheet")

	if filePath == "" || sheetName == "" || cellRange == "" {
		return nil, fmt.Errorf("file_path, sheet_name, and range are required")
	}

	values, ok := args["values"]
	if !ok {
		return nil, fmt.Errorf("values is required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Create new sheet if requested
	if newSheet {
		_, err := f.NewSheet(sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to create sheet: %w", err)
		}
	}

	// Parse starting cell from range
	parts := strings.Split(cellRange, ":")
	startCol, startRow, err := excelize.CellNameToCoordinates(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid range: %w", err)
	}

	// Write values
	cellsWritten := 0
	if rows, ok := values.([]interface{}); ok {
		for rowIdx, row := range rows {
			if rowData, ok := row.([]interface{}); ok {
				for colIdx, val := range rowData {
					cellName, _ := excelize.CoordinatesToCellName(startCol+colIdx, startRow+rowIdx)
					if err := f.SetCellValue(sheetName, cellName, val); err != nil {
						return nil, fmt.Errorf("failed to write cell %s: %w", cellName, err)
					}
					cellsWritten++
				}
			}
		}
	}

	if err := f.Save(); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return map[string]interface{}{
		"success":       true,
		"cells_written": cellsWritten,
		"range":         cellRange,
	}, nil
}

// createSheet creates a new sheet
func createSheet(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	sheetName := getString(args, "sheet_name")

	if filePath == "" || sheetName == "" {
		return nil, fmt.Errorf("file_path and sheet_name are required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}

	if err := f.Save(); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"index":   index,
		"name":    sheetName,
	}, nil
}

// deleteSheet deletes a sheet
func deleteSheet(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	sheetName := getString(args, "sheet_name")

	if filePath == "" || sheetName == "" {
		return nil, fmt.Errorf("file_path and sheet_name are required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if err := f.DeleteSheet(sheetName); err != nil {
		return nil, fmt.Errorf("failed to delete sheet: %w", err)
	}

	if err := f.Save(); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"deleted": sheetName,
	}, nil
}

// copySheet copies a sheet to a new sheet
func copySheet(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	srcSheet := getString(args, "src_sheet")
	dstSheet := getString(args, "dst_sheet")

	if filePath == "" || srcSheet == "" || dstSheet == "" {
		return nil, fmt.Errorf("file_path, src_sheet, and dst_sheet are required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	srcIndex, err := f.GetSheetIndex(srcSheet)
	if err != nil {
		return nil, fmt.Errorf("source sheet not found: %w", err)
	}

	dstIndex, err := f.NewSheet(dstSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination sheet: %w", err)
	}

	if err := f.CopySheet(srcIndex, dstIndex); err != nil {
		return nil, fmt.Errorf("failed to copy sheet: %w", err)
	}

	if err := f.Save(); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"src_sheet": srcSheet,
		"dst_sheet": dstSheet,
	}, nil
}

// getCell gets the value of a specific cell
func getCell(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	sheetName := getString(args, "sheet_name")
	cell := getString(args, "cell")

	if filePath == "" || sheetName == "" || cell == "" {
		return nil, fmt.Errorf("file_path, sheet_name, and cell are required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	value, err := f.GetCellValue(sheetName, cell)
	if err != nil {
		return nil, fmt.Errorf("failed to get cell: %w", err)
	}

	formula, _ := f.GetCellFormula(sheetName, cell)

	return map[string]interface{}{
		"cell":    cell,
		"value":   value,
		"formula": formula,
	}, nil
}

// setCell sets the value of a specific cell
func setCell(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	sheetName := getString(args, "sheet_name")
	cell := getString(args, "cell")
	value := args["value"]

	if filePath == "" || sheetName == "" || cell == "" {
		return nil, fmt.Errorf("file_path, sheet_name, and cell are required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Check if value is a formula
	if strVal, ok := value.(string); ok && strings.HasPrefix(strVal, "=") {
		if err := f.SetCellFormula(sheetName, cell, strVal); err != nil {
			return nil, fmt.Errorf("failed to set formula: %w", err)
		}
	} else {
		if err := f.SetCellValue(sheetName, cell, value); err != nil {
			return nil, fmt.Errorf("failed to set cell: %w", err)
		}
	}

	if err := f.Save(); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"cell":    cell,
		"value":   value,
	}, nil
}

// setFormula sets a formula in a specific cell
func setFormula(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")
	sheetName := getString(args, "sheet_name")
	cell := getString(args, "cell")
	formula := getString(args, "formula")

	if filePath == "" || sheetName == "" || cell == "" || formula == "" {
		return nil, fmt.Errorf("file_path, sheet_name, cell, and formula are required")
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if err := f.SetCellFormula(sheetName, cell, formula); err != nil {
		return nil, fmt.Errorf("failed to set formula: %w", err)
	}

	if err := f.Save(); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"cell":    cell,
		"formula": formula,
	}, nil
}

// createWorkbook creates a new Excel workbook
func createWorkbook(args map[string]interface{}) (interface{}, error) {
	filePath := getString(args, "file_path")

	if filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	f := excelize.NewFile()
	defer f.Close()

	if err := f.SaveAs(filePath); err != nil {
		return nil, fmt.Errorf("failed to create workbook: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"file_path": filePath,
		"sheets":    f.GetSheetList(),
	}, nil
}
