package sqlitetools

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// DB is the global database connection
var DB *sql.DB
var DBPath string

// InitDB initializes the database connection
func InitDB(path string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	DBPath = path
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func db_info(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Get file info
	fileInfo, err := os.Stat(DBPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Count tables
	var tableCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count tables: %w", err)
	}

	return map[string]interface{}{
		"path":        DBPath,
		"size_bytes":  fileSize,
		"table_count": tableCount,
	}, nil
}

func list_tables(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := DB.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	return map[string]interface{}{
		"tables": tables,
		"count":  len(tables),
	}, nil
}

func get_table_schema(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tableName, ok := args["tableName"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'tableName' not found or wrong type")
	}

	rows, err := DB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}
	defer rows.Close()

	var columns []map[string]interface{}
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}

		columns = append(columns, map[string]interface{}{
			"name":          name,
			"type":          colType,
			"not_null":      notNull == 1,
			"default_value": dfltValue,
			"primary_key":   pk == 1,
		})
	}

	return map[string]interface{}{
		"table_name": tableName,
		"columns":    columns,
	}, nil
}

func create_record(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	table, ok := args["table"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'table' not found or wrong type")
	}

	data, ok := args["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("required parameter 'data' not found or wrong type")
	}

	// Build INSERT query
	var columns []string
	var placeholders []string
	var values []interface{}

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := DB.Exec(query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert record: %w", err)
	}

	lastID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"success":       true,
		"last_id":       lastID,
		"rows_affected": rowsAffected,
	}, nil
}

func read_records(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	table, ok := args["table"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'table' not found or wrong type")
	}

	// Build SELECT query
	query := fmt.Sprintf("SELECT * FROM %s", table)
	var queryArgs []interface{}

	// Add conditions
	if conditions, ok := args["conditions"].(map[string]interface{}); ok && len(conditions) > 0 {
		var whereClauses []string
		for col, val := range conditions {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
			queryArgs = append(queryArgs, val)
		}
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Add limit
	if limit, ok := args["limit"].(float64); ok {
		query += fmt.Sprintf(" LIMIT %d", int(limit))
	}

	// Add offset
	if offset, ok := args["offset"].(float64); ok {
		query += fmt.Sprintf(" OFFSET %d", int(offset))
	}

	rows, err := DB.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to read records: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create a map for this row
		record := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle []byte values
			if b, ok := val.([]byte); ok {
				record[col] = string(b)
			} else {
				record[col] = val
			}
		}
		records = append(records, record)
	}

	return map[string]interface{}{
		"records": records,
		"count":   len(records),
	}, nil
}

func update_records(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	table, ok := args["table"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'table' not found or wrong type")
	}

	data, ok := args["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("required parameter 'data' not found or wrong type")
	}

	conditions, ok := args["conditions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("required parameter 'conditions' not found or wrong type")
	}

	// Build UPDATE query
	var setClauses []string
	var values []interface{}

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	var whereClauses []string
	for col, val := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		table,
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "))

	result, err := DB.Exec(query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to update records: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"success":       true,
		"rows_affected": rowsAffected,
	}, nil
}

func delete_records(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	table, ok := args["table"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'table' not found or wrong type")
	}

	conditions, ok := args["conditions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("required parameter 'conditions' not found or wrong type")
	}

	// Build DELETE query
	var whereClauses []string
	var values []interface{}

	for col, val := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		table,
		strings.Join(whereClauses, " AND "))

	result, err := DB.Exec(query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to delete records: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"success":       true,
		"rows_affected": rowsAffected,
	}, nil
}

func query_exec(args map[string]interface{}) (interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlQuery, ok := args["sql"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'sql' not found or wrong type")
	}

	// Get optional values
	var queryValues []interface{}
	if values, ok := args["values"].([]interface{}); ok {
		queryValues = values
	}

	// Determine if it's a SELECT or other query
	trimmedSQL := strings.TrimSpace(strings.ToUpper(sqlQuery))
	isSelect := strings.HasPrefix(trimmedSQL, "SELECT") ||
	            strings.HasPrefix(trimmedSQL, "PRAGMA") ||
	            strings.HasPrefix(trimmedSQL, "EXPLAIN")

	if isSelect {
		rows, err := DB.Query(sqlQuery, queryValues...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		var records []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			record := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					record[col] = string(b)
				} else {
					record[col] = val
				}
			}
			records = append(records, record)
		}

		return map[string]interface{}{
			"records": records,
			"count":   len(records),
		}, nil
	} else {
		result, err := DB.Exec(sqlQuery, queryValues...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}

		lastID, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()

		return map[string]interface{}{
			"success":       true,
			"last_id":       lastID,
			"rows_affected": rowsAffected,
		}, nil
	}
}

// ToJSON converts a result to JSON string
func ToJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
