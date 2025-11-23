package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/imran31415/godemode/pkg/validator"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// InterpreterExecutor runs Go code directly using yaegi interpreter
// This eliminates the WASM compilation overhead (~2-3s) completely
type InterpreterExecutor struct {
	validator       *validator.Validator
	defaultTimeout  time.Duration
	interpreterPool *interpreterPool
	customSymbols   map[string]map[string]interface{}
}

// interpreterPool maintains a pool of reusable yaegi interpreters
type interpreterPool struct {
	pool chan *interp.Interpreter
	mu   sync.Mutex
}

// NewInterpreterExecutor creates a new interpreter-based executor
func NewInterpreterExecutor() *InterpreterExecutor {
	return &InterpreterExecutor{
		validator:      validator.NewValidator(),
		defaultTimeout: 30 * time.Second,
		interpreterPool: newInterpreterPool(5), // Pool of 5 interpreters
	}
}

// newInterpreterPool creates a pool of pre-initialized interpreters
func newInterpreterPool(size int) *interpreterPool {
	pool := &interpreterPool{
		pool: make(chan *interp.Interpreter, size),
	}

	// Pre-create interpreters
	for i := 0; i < size; i++ {
		i := interp.New(interp.Options{})
		i.Use(stdlib.Symbols) // Load standard library
		pool.pool <- i
	}

	return pool
}

// get retrieves an interpreter from the pool
func (p *interpreterPool) get() *interp.Interpreter {
	select {
	case i := <-p.pool:
		return i
	default:
		// Pool empty, create new interpreter
		i := interp.New(interp.Options{})
		i.Use(stdlib.Symbols)
		return i
	}
}

// put returns an interpreter to the pool
func (p *interpreterPool) put(i *interp.Interpreter) {
	select {
	case p.pool <- i:
		// Successfully returned to pool
	default:
		// Pool full, let it be garbage collected
	}
}

// Execute interprets and runs Go source code directly (no compilation)
func (e *InterpreterExecutor) Execute(ctx context.Context, sourceCode string, timeout time.Duration) (*ExecutionResult, error) {
	startTime := time.Now()

	// Use default timeout if none specified
	if timeout == 0 {
		timeout = e.defaultTimeout
	}

	// Step 1: Validate source code
	if err := e.validator.Validate(sourceCode); err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("validation failed: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	// Step 2: Execute directly (no compilation needed!)
	result, err := e.executeInterpreted(ctx, sourceCode, timeout)
	result.Duration = time.Since(startTime)

	return result, err
}

// executeInterpreted runs Go code using yaegi interpreter
func (e *InterpreterExecutor) executeInterpreted(ctx context.Context, sourceCode string, timeout time.Duration) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Success: false,
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get interpreter from pool
	i := e.interpreterPool.get()
	defer e.interpreterPool.put(i)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	// Create pipes for output capture
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Channel to signal completion
	done := make(chan error, 1)

	// Run interpreter in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()

		// Evaluate the code
		_, err := i.Eval(sourceCode)
		done <- err
	}()

	// Copy output
	go func() {
		io.Copy(&stdout, stdoutR)
	}()
	go func() {
		io.Copy(&stderr, stderrR)
	}()

	// Wait for completion or timeout
	var err error
	select {
	case err = <-done:
		// Execution completed
	case <-execCtx.Done():
		err = execCtx.Err()
	}

	// Restore stdout/stderr
	stdoutW.Close()
	stderrW.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	// Wait a bit for output to flush
	time.Sleep(10 * time.Millisecond)

	stdoutR.Close()
	stderrR.Close()

	// Capture output
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Handle execution errors
	if err != nil {
		execErr := e.classifyExecutionError(err)
		result.Error = execErr.Message
		return result, execErr
	}

	result.Success = true
	return result, nil
}

// classifyExecutionError categorizes execution errors
func (e *InterpreterExecutor) classifyExecutionError(err error) *ExecutionError {
	errMsg := err.Error()

	// Check for timeout
	if err == context.DeadlineExceeded {
		return &ExecutionError{
			Type:    "timeout",
			Message: "execution timeout exceeded",
		}
	}

	// Check for context cancellation
	if err == context.Canceled {
		return &ExecutionError{
			Type:    "canceled",
			Message: "execution was canceled",
		}
	}

	// Check for panic
	if strings.Contains(errMsg, "panic") {
		return &ExecutionError{
			Type:    "panic",
			Message: errMsg,
		}
	}

	// Generic execution error
	return &ExecutionError{
		Type:    "runtime",
		Message: errMsg,
	}
}

// ExecuteSimple is a convenience method
func (e *InterpreterExecutor) ExecuteSimple(sourceCode string) (*ExecutionResult, error) {
	return e.Execute(context.Background(), sourceCode, e.defaultTimeout)
}

// SetDefaultTimeout sets the default execution timeout
func (e *InterpreterExecutor) SetDefaultTimeout(timeout time.Duration) {
	e.defaultTimeout = timeout
}

// SetCustomSymbols sets custom symbols that will be available to executed code
// The symbols map should be: package path -> symbol name -> value
// Example: map[string]map[string]interface{}{"main/main": {"myFunc": myFunction}}
func (e *InterpreterExecutor) SetCustomSymbols(symbols map[string]map[string]interface{}) {
	e.customSymbols = symbols
}

// ExecuteWithSymbols executes code with custom symbols injected
// This is useful for providing tool registries or other external dependencies
func (e *InterpreterExecutor) ExecuteWithSymbols(ctx context.Context, sourceCode string, timeout time.Duration, symbols map[string]map[string]interface{}) (*ExecutionResult, error) {
	startTime := time.Now()

	if timeout == 0 {
		timeout = e.defaultTimeout
	}

	// Skip validation for code with external dependencies
	// The validator may not understand custom symbols

	// Execute with custom symbols
	result, err := e.executeWithCustomSymbols(ctx, sourceCode, timeout, symbols)
	result.Duration = time.Since(startTime)

	return result, err
}

// ExecuteGeneratedCode is a high-level API for executing LLM-generated code
// It handles markdown extraction, code preprocessing, and registry injection
func (e *InterpreterExecutor) ExecuteGeneratedCode(ctx context.Context, rawCode string, timeout time.Duration, registryCall func(string, map[string]interface{}) (interface{}, error)) (*ExecutionResult, error) {
	// Preprocess the code
	preprocessor := NewCodePreprocessor()
	processedCode := preprocessor.Process(rawCode, "registryCall")

	// Validate basic structure
	if err := preprocessor.ValidateBasicStructure(processedCode); err != "" {
		return &ExecutionResult{
			Success: false,
			Error:   "code validation failed: " + err,
		}, nil
	}

	// Create symbols map with the registry call function
	symbols := map[string]map[string]interface{}{
		"main/main": {
			"registryCall": registryCall,
		},
	}

	// Execute with the injected symbols
	return e.ExecuteWithSymbols(ctx, processedCode, timeout, symbols)
}

// executeWithCustomSymbols runs code with custom symbols injected
func (e *InterpreterExecutor) executeWithCustomSymbols(ctx context.Context, sourceCode string, timeout time.Duration, symbols map[string]map[string]interface{}) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Success: false,
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create fresh interpreter (don't use pool for custom symbols)
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	// Inject custom symbols
	if symbols != nil {
		reflectSymbols := make(map[string]map[string]reflect.Value)
		for pkg, syms := range symbols {
			reflectSymbols[pkg] = make(map[string]reflect.Value)
			for name, val := range syms {
				reflectSymbols[pkg][name] = reflect.ValueOf(val)
			}
		}
		i.Use(reflectSymbols)
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()

		_, err := i.Eval(sourceCode)
		done <- err
	}()

	go func() {
		io.Copy(&stdout, stdoutR)
	}()
	go func() {
		io.Copy(&stderr, stderrR)
	}()

	var err error
	select {
	case err = <-done:
	case <-execCtx.Done():
		err = execCtx.Err()
	}

	stdoutW.Close()
	stderrW.Close()
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	time.Sleep(10 * time.Millisecond)

	stdoutR.Close()
	stderrR.Close()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		execErr := e.classifyExecutionError(err)
		result.Error = execErr.Message
		return result, execErr
	}

	result.Success = true
	return result, nil
}
