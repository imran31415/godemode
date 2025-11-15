package fstools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Generated tool implementations with real filesystem operations

func readFile(args map[string]interface{}) (interface{}, error) {
	// Required parameter: path (string)
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'path' not found or wrong type")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"content": string(content),
		"size":    len(content),
	}, nil
}

func writeFile(args map[string]interface{}) (interface{}, error) {
	// Required parameter: path (string)
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'path' not found or wrong type")
	}

	// Required parameter: content (string)
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'content' not found or wrong type")
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"status":      "success",
		"path":        path,
		"bytes_wrote": len(content),
	}, nil
}

func listDirectory(args map[string]interface{}) (interface{}, error) {
	// Required parameter: path (string)
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'path' not found or wrong type")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	files := []map[string]interface{}{}
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, map[string]interface{}{
			"name":  entry.Name(),
			"isDir": entry.IsDir(),
			"size":  info.Size(),
		})
	}

	return map[string]interface{}{
		"path":  path,
		"files": files,
		"count": len(files),
	}, nil
}

func createDirectory(args map[string]interface{}) (interface{}, error) {
	// Required parameter: path (string)
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'path' not found or wrong type")
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return map[string]interface{}{
		"status": "success",
		"path":   path,
	}, nil
}

func deleteFile(args map[string]interface{}) (interface{}, error) {
	// Required parameter: path (string)
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'path' not found or wrong type")
	}

	err := os.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"status": "success",
		"path":   path,
	}, nil
}

func getFileInfo(args map[string]interface{}) (interface{}, error) {
	// Required parameter: path (string)
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'path' not found or wrong type")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return map[string]interface{}{
		"path":     path,
		"size":     info.Size(),
		"isDir":    info.IsDir(),
		"modified": info.ModTime().Format("2006-01-02 15:04:05"),
		"mode":     info.Mode().String(),
	}, nil
}

func searchFiles(args map[string]interface{}) (interface{}, error) {
	// Required parameter: directory (string)
	directory, ok := args["directory"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'directory' not found or wrong type")
	}

	// Required parameter: pattern (string)
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter 'pattern' not found or wrong type")
	}

	var matches []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			matches = append(matches, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	// Convert paths to relative if possible
	relMatches := []string{}
	for _, match := range matches {
		rel, err := filepath.Rel(directory, match)
		if err == nil {
			relMatches = append(relMatches, rel)
		} else {
			relMatches = append(relMatches, match)
		}
	}

	return map[string]interface{}{
		"directory": directory,
		"pattern":   pattern,
		"matches":   strings.Join(relMatches, ", "),
		"count":     len(matches),
	}, nil
}
