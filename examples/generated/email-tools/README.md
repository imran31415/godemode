# Generated GoDeMode Tools

This package was auto-generated from a mcp specification.

## Tools (3 total)

### sendEmail

Send an email to a recipient

**Parameters:**

- `body` (string) *(required)* - Email body content
- `cc` ([]interface{}) - CC recipients
- `to` (string) *(required)* - Recipient email address
- `subject` (string) *(required)* - Email subject line

### readEmail

Read an email by ID

**Parameters:**

- `emailId` (string) *(required)* - Unique email identifier

### listEmails

List emails with optional filters

**Parameters:**

- `folder` (string) - Email folder (inbox, sent, trash)
- `limit` (int) - Maximum number of emails to return
- `unreadOnly` (bool) - Only return unread emails

## Usage

```go
package main

import (
	"generated/emailtools"
)

func main() {
	// Create registry
	registry := emailtools.NewRegistry()

	// Call a tool
	result, err := registry.Call("toolName", map[string]interface{}{
		"param1": "value1",
	})
}
```
