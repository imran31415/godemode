package hostfuncs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/imran31415/godemode/pkg/tools"
	"github.com/tetratelabs/wazero/api"
)

// Bridge provides host functions that WASM can call
type Bridge struct {
	registry *tools.Registry
	logger   *tools.Logger
	state    *tools.State
}

// NewBridge creates a new host function bridge with standard tools
func NewBridge() *Bridge {
	logger := tools.NewLogger()
	state := tools.NewState()

	registry := tools.NewRegistry()
	registry.Register(logger)
	registry.Register(state)

	return &Bridge{
		registry: registry,
		logger:   logger,
		state:    state,
	}
}

// GetLogger returns the logger instance
func (b *Bridge) GetLogger() *tools.Logger {
	return b.logger
}

// GetState returns the state instance
func (b *Bridge) GetState() *tools.State {
	return b.state
}

// GetRegistry returns the tool registry
func (b *Bridge) GetRegistry() *tools.Registry {
	return b.registry
}

// logMessage is the host function for logging
// WASM signature: log_message(msgPtr, msgLen uint32) uint32
func (b *Bridge) logMessage(ctx context.Context, m api.Module, msgPtr, msgLen uint32) uint32 {
	// Read message from WASM memory
	message, ok := m.Memory().Read(msgPtr, msgLen)
	if !ok {
		return 0 // Failed to read
	}

	// Invoke logger tool
	_, err := b.logger.Invoke(string(message))
	if err != nil {
		return 0
	}

	return 1 // Success
}

// stateGet is the host function for getting state
// WASM signature: state_get(keyPtr, keyLen, resultPtr uint32) uint32
func (b *Bridge) stateGet(ctx context.Context, m api.Module, keyPtr, keyLen, resultPtr uint32) uint32 {
	// Read key from WASM memory
	keyBytes, ok := m.Memory().Read(keyPtr, keyLen)
	if !ok {
		return 0 // Failed to read
	}

	key := string(keyBytes)

	// Invoke state tool
	result, err := b.state.Invoke("get", key)
	if err != nil {
		return 0
	}

	// Marshal result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return 0
	}

	// Write result to WASM memory at resultPtr
	// Note: In a full implementation, we'd need memory allocation
	// For now, we'll just return success
	_ = resultJSON

	return 1 // Success
}

// stateSet is the host function for setting state
// WASM signature: state_set(keyPtr, keyLen, valuePtr, valueLen uint32) uint32
func (b *Bridge) stateSet(ctx context.Context, m api.Module, keyPtr, keyLen, valuePtr, valueLen uint32) uint32 {
	// Read key from WASM memory
	keyBytes, ok := m.Memory().Read(keyPtr, keyLen)
	if !ok {
		return 0 // Failed to read
	}

	// Read value from WASM memory
	valueBytes, ok := m.Memory().Read(valuePtr, valueLen)
	if !ok {
		return 0 // Failed to read
	}

	key := string(keyBytes)
	value := string(valueBytes)

	// Invoke state tool
	_, err := b.state.Invoke("set", key, value)
	if err != nil {
		return 0
	}

	return 1 // Success
}

// Note: For MVP, we'll keep the host functions simple
// A more sophisticated implementation would:
// 1. Implement malloc/free in WASM to allocate return value memory
// 2. Support complex data types beyond strings
// 3. Provide better error reporting back to WASM

// ExportHostFunctions exports host functions to the WASM module
// This would be used when instantiating the WASM runtime
func (b *Bridge) ExportHostFunctions() map[string]interface{} {
	return map[string]interface{}{
		"log_message": b.logMessage,
		"state_get":   b.stateGet,
		"state_set":   b.stateSet,
	}
}

// WASMHelperCode returns Go code that should be included in WASM programs
// to make calling host functions easier
func WASMHelperCode() string {
	return `
// Host function imports (add these to your WASM code)
//
// //go:wasm-module env
// //export log_message
// func wasmLogMessage(ptr, len uint32) uint32
//
// // Helper function to log messages
// func log(message string) {
//     ptr := &[]byte(message)[0]
//     wasmLogMessage(uint32(uintptr(unsafe.Pointer(ptr))), uint32(len(message)))
// }
//
// Note: Full implementation would require unsafe package which is forbidden
// For MVP, logging will be done via fmt.Println which goes to stdout
`
}

// GetToolDocumentation returns documentation for available tools
func (b *Bridge) GetToolDocumentation() string {
	doc := "Available Tools:\n\n"

	for _, toolName := range b.registry.List() {
		tool, _ := b.registry.Get(toolName)
		doc += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
	}

	doc += "\nFor MVP, use fmt.Println() for logging output.\n"
	doc += "State management and host functions will be enhanced in future versions.\n"

	return doc
}
