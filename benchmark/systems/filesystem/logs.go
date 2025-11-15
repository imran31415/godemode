package filesystem

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Raw       string
}

// LogSystem handles reading log files
type LogSystem struct {
	logsPath string
}

// NewLogSystem creates a new log system
func NewLogSystem(logsPath string) *LogSystem {
	return &LogSystem{logsPath: logsPath}
}

// SearchLogs searches for a pattern in log files
func (ls *LogSystem) SearchLogs(pattern string, timeWindow time.Duration) ([]*LogEntry, error) {
	// Get all log files
	files, err := filepath.Glob(filepath.Join(ls.logsPath, "*.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to list log files: %w", err)
	}

	var entries []*LogEntry
	cutoffTime := time.Now().Add(-timeWindow)

	for _, file := range files {
		fileEntries, err := ls.searchFile(file, pattern, cutoffTime)
		if err != nil {
			continue  // Skip files with errors
		}
		entries = append(entries, fileEntries...)
	}

	return entries, nil
}

// searchFile searches a single log file for pattern
func (ls *LogSystem) searchFile(filename, pattern string, cutoffTime time.Time) ([]*LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []*LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line matches pattern (simple string contains)
		if !strings.Contains(line, pattern) {
			continue
		}

		// Parse log entry
		entry := ls.parseLogLine(line)

		// Filter by time window
		if !entry.Timestamp.IsZero() && entry.Timestamp.Before(cutoffTime) {
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// parseLogLine parses a log line into a LogEntry
// Supports common formats: [TIMESTAMP] LEVEL: Message
func (ls *LogSystem) parseLogLine(line string) *LogEntry {
	entry := &LogEntry{
		Raw: line,
	}

	// Try to extract timestamp (e.g., 2024-01-01 14:23:45)
	timestampPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}[\sT]\d{2}:\d{2}:\d{2})`)
	if match := timestampPattern.FindStringSubmatch(line); len(match) > 1 {
		ts, err := time.Parse("2006-01-02 15:04:05", match[1])
		if err == nil {
			entry.Timestamp = ts
		}
	}

	// Try to extract level (ERROR, WARN, INFO, DEBUG)
	levelPattern := regexp.MustCompile(`\b(ERROR|WARN|WARNING|INFO|DEBUG)\b`)
	if match := levelPattern.FindStringSubmatch(line); len(match) > 1 {
		entry.Level = match[1]
	}

	// Extract message (everything after level, or full line)
	if entry.Level != "" {
		parts := strings.SplitN(line, entry.Level, 2)
		if len(parts) == 2 {
			entry.Message = strings.TrimSpace(parts[1])
			entry.Message = strings.TrimPrefix(entry.Message, ":")
			entry.Message = strings.TrimSpace(entry.Message)
		}
	} else {
		entry.Message = line
	}

	return entry
}

// ExtractErrorContext extracts lines around an error code for context
func (ls *LogSystem) ExtractErrorContext(errorCode string, contextLines int) (string, error) {
	files, err := filepath.Glob(filepath.Join(ls.logsPath, "*.log"))
	if err != nil {
		return "", fmt.Errorf("failed to list log files: %w", err)
	}

	for _, file := range files {
		context, err := ls.extractFromFile(file, errorCode, contextLines)
		if err == nil && context != "" {
			return context, nil
		}
	}

	return "", fmt.Errorf("error code not found in logs: %s", errorCode)
}

// extractFromFile extracts context from a single file
func (ls *LogSystem) extractFromFile(filename, errorCode string, contextLines int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Read all lines
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Find the error code
	for i, line := range lines {
		if strings.Contains(line, errorCode) {
			// Extract context
			start := i - contextLines
			if start < 0 {
				start = 0
			}

			end := i + contextLines + 1
			if end > len(lines) {
				end = len(lines)
			}

			contextLines := lines[start:end]
			return strings.Join(contextLines, "\n"), nil
		}
	}

	return "", nil
}

// WriteLog appends a log entry (for testing/simulation)
func (ls *LogSystem) WriteLog(filename, level, message string) error {
	fullPath := filepath.Join(ls.logsPath, filename)

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, message)

	if _, err := file.WriteString(logLine); err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// Reset clears all log files
func (ls *LogSystem) Reset() error {
	files, err := filepath.Glob(filepath.Join(ls.logsPath, "*.log"))
	if err != nil {
		return err
	}

	for _, file := range files {
		os.Remove(file)
	}

	return nil
}
