package scenarios

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/email"
	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/systems/security"
)

// IsolatedTestEnv wraps a TestEnvironment with cleanup capabilities
type IsolatedTestEnv struct {
	*TestEnvironment
	TempDir string
}

// Cleanup removes all temporary data created for this test
func (e *IsolatedTestEnv) Cleanup() error {
	if e.TempDir != "" {
		return os.RemoveAll(e.TempDir)
	}
	return nil
}

// NewIsolatedTestEnvironment creates a completely isolated test environment
// Each test gets its own copy of fixture data in a temp directory
func NewIsolatedTestEnvironment(fixturesPath string) (*IsolatedTestEnv, error) {
	// Create temp directory for this test
	tempDir, err := os.MkdirTemp("", "godemode-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Copy fixtures to temp directory
	if err := copyDir(fixturesPath, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to copy fixtures: %w", err)
	}

	// Create test environment using isolated temp directory
	env, err := createTestEnvironment(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, err
	}

	return &IsolatedTestEnv{
		TestEnvironment: env,
		TempDir:         tempDir,
	}, nil
}

// NewIsolatedTestEnvironmentWithSeparateGraph creates an isolated environment
// but uses a separate temp directory for the graph database (for process isolation)
func NewIsolatedTestEnvironmentWithSeparateGraph(fixturesPath string) (*IsolatedTestEnv, error) {
	// Create temp directory for fixtures
	tempDir, err := os.MkdirTemp("", "godemode-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Copy fixtures to temp directory
	if err := copyDir(fixturesPath, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to copy fixtures: %w", err)
	}

	// Create separate temp directory for graph database
	graphDir, err := os.MkdirTemp("", "godemode-graph-*")
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create graph directory: %w", err)
	}

	// Copy graph data to separate directory
	if err := copyDir(filepath.Join(fixturesPath, "graph_data"), graphDir); err != nil {
		os.RemoveAll(tempDir)
		os.RemoveAll(graphDir)
		return nil, fmt.Errorf("failed to copy graph data: %w", err)
	}

	// Create test environment with custom graph path
	emailSystem := email.NewEmailSystem(
		filepath.Join(tempDir, "emails"),
		filepath.Join(tempDir, "emails", "sent"),
	)

	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		os.RemoveAll(tempDir)
		os.RemoveAll(graphDir)
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	graphDB, err := graph.NewKnowledgeGraph(graphDir)
	if err != nil {
		os.RemoveAll(tempDir)
		os.RemoveAll(graphDir)
		return nil, fmt.Errorf("failed to create knowledge graph: %w", err)
	}

	logSystem := filesystem.NewLogSystem(filepath.Join(tempDir, "logs"))
	configSystem := filesystem.NewConfigSystem(filepath.Join(tempDir, "configs"))
	securityMonitor := security.NewSecurityMonitor()

	env := &TestEnvironment{
		EmailSystem:     emailSystem,
		Database:        db,
		Graph:           graphDB,
		LogSystem:       logSystem,
		ConfigSystem:    configSystem,
		SecurityMonitor: securityMonitor,
	}

	// Note: We need to cleanup both tempDir and graphDir
	isolatedEnv := &IsolatedTestEnv{
		TestEnvironment: env,
		TempDir:         tempDir,
	}

	// Store graphDir for cleanup (extend cleanup to handle both)
	// For now, we'll add graphDir to the cleanup manually where used

	return isolatedEnv, nil
}

// createTestEnvironment creates a TestEnvironment from a fixtures directory
func createTestEnvironment(fixturesPath string) (*TestEnvironment, error) {
	emailSystem := email.NewEmailSystem(
		filepath.Join(fixturesPath, "emails"),
		filepath.Join(fixturesPath, "emails", "sent"),
	)

	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	graphDB, err := graph.NewKnowledgeGraph(filepath.Join(fixturesPath, "graph_data"))
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge graph: %w", err)
	}

	logSystem := filesystem.NewLogSystem(filepath.Join(fixturesPath, "logs"))
	configSystem := filesystem.NewConfigSystem(filepath.Join(fixturesPath, "configs"))
	securityMonitor := security.NewSecurityMonitor()

	return &TestEnvironment{
		EmailSystem:     emailSystem,
		Database:        db,
		Graph:           graphDB,
		LogSystem:       logSystem,
		ConfigSystem:    configSystem,
		SecurityMonitor: securityMonitor,
	}, nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
