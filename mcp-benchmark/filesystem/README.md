# Generated GoDeMode Tools

This package was auto-generated from a mcp specification.

## Tools (7 total)

### readFile

Read contents of a file

**Parameters:**

- `path` (string) *(required)* - File path to read

### writeFile

Write content to a file

**Parameters:**

- `path` (string) *(required)* - File path to write
- `content` (string) *(required)* - Content to write

### listDirectory

List files and directories in a path

**Parameters:**

- `path` (string) *(required)* - Directory path to list

### createDirectory

Create a new directory

**Parameters:**

- `path` (string) *(required)* - Directory path to create

### deleteFile

Delete a file

**Parameters:**

- `path` (string) *(required)* - File path to delete

### getFileInfo

Get file metadata (size, modified time, etc.)

**Parameters:**

- `path` (string) *(required)* - File path

### searchFiles

Search for files matching a pattern

**Parameters:**

- `directory` (string) *(required)* - Directory to search in
- `pattern` (string) *(required)* - File pattern to match (e.g., '*.txt')

## Usage

```go
package main

import (
	"generated/fstools"
)

func main() {
	// Create registry
	registry := fstools.NewRegistry()

	// Call a tool
	result, err := registry.Call("toolName", map[string]interface{}{
		"param1": "value1",
	})
}
```
