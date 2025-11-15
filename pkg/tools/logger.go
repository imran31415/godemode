package tools

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
}

// Logger is a tool that allows WASM code to log messages
type Logger struct {
	mu      sync.Mutex
	entries []LogEntry
}

// NewLogger creates a new Logger tool
func NewLogger() *Logger {
	return &Logger{
		entries: make([]LogEntry, 0),
	}
}

// Name returns the tool name
func (l *Logger) Name() string {
	return "log"
}

// Description returns the tool description
func (l *Logger) Description() string {
	return "Log messages from WASM code"
}

// Invoke logs a message
// Expected args: message string, optional level string
func (l *Logger) Invoke(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("log requires at least 1 argument (message)")
	}

	message, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("first argument must be a string")
	}

	level := "INFO"
	if len(args) > 1 {
		if lvl, ok := args[1].(string); ok {
			level = strings.ToUpper(lvl)
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}
	l.entries = append(l.entries, entry)

	return map[string]interface{}{
		"success": true,
		"logged":  message,
	}, nil
}

// GetEntries returns all logged entries
func (l *Logger) GetEntries() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Return a copy
	entries := make([]LogEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// Clear removes all log entries
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = make([]LogEntry, 0)
}

// String returns a formatted string of all log entries
func (l *Logger) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	var sb strings.Builder
	for _, entry := range l.entries {
		sb.WriteString(fmt.Sprintf("[%s] %s: %s\n",
			entry.Timestamp.Format("15:04:05"),
			entry.Level,
			entry.Message))
	}
	return sb.String()
}
