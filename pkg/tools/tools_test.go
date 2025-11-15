package tools

import (
	"testing"
)

func TestLogger(t *testing.T) {
	logger := NewLogger()

	// Test logging
	result, err := logger.Invoke("Test message")
	if err != nil {
		t.Errorf("Logging failed: %v", err)
	}

	if result == nil {
		t.Error("Should return result")
	}

	// Test entries
	entries := logger.GetEntries()
	if len(entries) != 1 {
		t.Errorf("Should have 1 entry, got %d", len(entries))
	}

	if entries[0].Message != "Test message" {
		t.Errorf("Message should be 'Test message', got '%s'", entries[0].Message)
	}

	if entries[0].Level != "INFO" {
		t.Errorf("Default level should be INFO, got %s", entries[0].Level)
	}

	// Test with custom level
	logger.Invoke("Error message", "ERROR")
	entries = logger.GetEntries()
	if len(entries) != 2 {
		t.Errorf("Should have 2 entries, got %d", len(entries))
	}

	if entries[1].Level != "ERROR" {
		t.Errorf("Level should be ERROR, got %s", entries[1].Level)
	}

	// Test clear
	logger.Clear()
	if len(logger.GetEntries()) != 0 {
		t.Error("Should have no entries after clear")
	}
}

func TestState(t *testing.T) {
	state := NewState()

	// Test set
	result, err := state.Invoke("set", "key1", "value1")
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}
	if result == nil {
		t.Error("Should return result")
	}

	// Test get
	result, err = state.Invoke("get", "key1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Result should be a map")
	}

	if !resultMap["exists"].(bool) {
		t.Error("Key should exist")
	}

	if resultMap["value"] != "value1" {
		t.Errorf("Value should be 'value1', got %v", resultMap["value"])
	}

	// Test has
	result, err = state.Invoke("has", "key1")
	if err != nil {
		t.Errorf("Has failed: %v", err)
	}

	resultMap, _ = result.(map[string]interface{})
	if !resultMap["exists"].(bool) {
		t.Error("Key should exist")
	}

	// Test delete
	_, err = state.Invoke("delete", "key1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	result, _ = state.Invoke("has", "key1")
	resultMap, _ = result.(map[string]interface{})
	if resultMap["exists"].(bool) {
		t.Error("Key should not exist after delete")
	}

	// Test size
	state.Invoke("set", "a", 1)
	state.Invoke("set", "b", 2)
	if state.Size() != 2 {
		t.Errorf("Size should be 2, got %d", state.Size())
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()
	logger := NewLogger()

	// Test register
	err := registry.Register(logger)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// Test duplicate register
	err = registry.Register(logger)
	if err == nil {
		t.Error("Should fail to register duplicate")
	}

	// Test get
	tool, found := registry.Get("log")
	if !found {
		t.Error("Should find registered tool")
	}

	if tool.Name() != "log" {
		t.Errorf("Tool name should be 'log', got '%s'", tool.Name())
	}

	// Test list
	names := registry.List()
	if len(names) != 1 {
		t.Errorf("Should have 1 tool, got %d", len(names))
	}

	// Test unregister
	registry.Unregister("log")
	_, found = registry.Get("log")
	if found {
		t.Error("Should not find unregistered tool")
	}
}
