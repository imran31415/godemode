# Generated GoDeMode Tools

This package was auto-generated from a mcp specification.

## Tools (5 total)

### add

Add two numbers together

**Parameters:**

- `a` (float64) *(required)* - First number
- `b` (float64) *(required)* - Second number

### getCurrentTime

Get the current time in RFC3339 format

### generateUUID

Generate a random UUID

### concatenateStrings

Concatenate an array of strings with a separator

**Parameters:**

- `strings` ([]interface{}) *(required)* - Array of strings to concatenate
- `separator` (string) - Separator to use between strings

### reverseString

Reverse a string

**Parameters:**

- `text` (string) *(required)* - String to reverse

## Usage

```go
package main

import (
	"generated/utilitytools"
)

func main() {
	// Create registry
	registry := utilitytools.NewRegistry()

	// Call a tool
	result, err := registry.Call("toolName", map[string]interface{}{
		"param1": "value1",
	})
}
```
