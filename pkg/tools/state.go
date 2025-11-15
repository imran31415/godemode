package tools

import (
	"fmt"
	"sync"
)

// State is a simple key-value store tool for WASM code
type State struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewState creates a new State tool
func NewState() *State {
	return &State{
		data: make(map[string]interface{}),
	}
}

// Name returns the tool name
func (s *State) Name() string {
	return "state"
}

// Description returns the tool description
func (s *State) Description() string {
	return "Key-value state management for WASM code"
}

// Invoke handles state operations
// Operations: "get", "set", "delete", "has", "keys", "clear"
// Args: operation string, key string, optional value
func (s *State) Invoke(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("state requires at least 2 arguments (operation, key)")
	}

	operation, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("first argument must be a string (operation)")
	}

	key, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("second argument must be a string (key)")
	}

	switch operation {
	case "get":
		return s.get(key)
	case "set":
		if len(args) < 3 {
			return nil, fmt.Errorf("set operation requires a value")
		}
		return s.set(key, args[2])
	case "delete":
		return s.delete(key)
	case "has":
		return s.has(key)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// get retrieves a value from state
func (s *State) get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return map[string]interface{}{
			"exists": false,
			"value":  nil,
		}, nil
	}

	return map[string]interface{}{
		"exists": true,
		"value":  value,
	}, nil
}

// set stores a value in state
func (s *State) set(key string, value interface{}) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value

	return map[string]interface{}{
		"success": true,
		"key":     key,
	}, nil
}

// delete removes a value from state
func (s *State) delete(key string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)

	return map[string]interface{}{
		"success": true,
		"deleted": key,
	}, nil
}

// has checks if a key exists in state
func (s *State) has(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.data[key]

	return map[string]interface{}{
		"exists": exists,
	}, nil
}

// Keys returns all keys in state
func (s *State) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys
}

// Clear removes all entries from state
func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]interface{})
}

// Size returns the number of entries in state
func (s *State) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}

// GetAll returns a copy of all state data
func (s *State) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	data := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		data[k] = v
	}
	return data
}
