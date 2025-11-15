# Generated GoDeMode Tools

This package was auto-generated from a openapi specification.

## Tools (4 total)

### listUsers

List all users

**Parameters:**

- `page` (int64) - Page number
- `limit` (int64) - Items per page

### createUser

Create a new user

**Parameters:**

- `username` (string) *(required)* - No description provided
- `email` (string) *(required)* - No description provided
- `password` (string) *(required)* - No description provided
- `firstName` (string) - No description provided
- `lastName` (string) - No description provided

### getUser

Get user by ID

**Parameters:**

- `id` (string) *(required)* - User ID

### deleteUser

Delete a user

**Parameters:**

- `id` (string) *(required)* - User ID

## Usage

```go
package main

import (
	"generated/userapi"
)

func main() {
	// Create registry
	registry := userapi.NewRegistry()

	// Call a tool
	result, err := registry.Call("toolName", map[string]interface{}{
		"param1": "value1",
	})
}
```
