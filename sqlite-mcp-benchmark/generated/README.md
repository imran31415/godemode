# Generated GoDeMode Tools

This package was auto-generated from a mcp specification.

## Tools (8 total)

### db_info

Get detailed information about the connected database including file path, size, and table count.

### list_tables

List all tables in the database.

### get_table_schema

Get detailed information about a table's schema including column names, types, and constraints.

**Parameters:**

- `tableName` (string) *(required)* - Name of the table to describe

### create_record

Insert a new record into a table.

**Parameters:**

- `table` (string) *(required)* - Name of the table
- `data` (map[string]interface{}) *(required)* - Record data as key-value pairs

### read_records

Query records from a table with optional filtering, pagination, and sorting.

**Parameters:**

- `table` (string) *(required)* - Name of the table
- `conditions` (map[string]interface{}) - Filter conditions as key-value pairs
- `limit` (float64) - Maximum number of records to return
- `offset` (float64) - Number of records to skip

### update_records

Update records in a table that match specified conditions.

**Parameters:**

- `table` (string) *(required)* - Name of the table
- `data` (map[string]interface{}) *(required)* - New values as key-value pairs
- `conditions` (map[string]interface{}) *(required)* - Filter conditions to select records to update

### delete_records

Delete records from a table that match specified conditions.

**Parameters:**

- `conditions` (map[string]interface{}) *(required)* - Filter conditions to select records to delete
- `table` (string) *(required)* - Name of the table

### query

Execute a custom SQL query against the connected SQLite database. Supports parameterized queries for security.

**Parameters:**

- `sql` (string) *(required)* - The SQL query to execute
- `values` ([]interface{}) - Array of parameter values to use in the query

## Usage

```go
package main

import (
	"generated/sqlitetools"
)

func main() {
	// Create registry
	registry := sqlitetools.NewRegistry()

	// Call a tool
	result, err := registry.Call("toolName", map[string]interface{}{
		"param1": "value1",
	})
}
```
