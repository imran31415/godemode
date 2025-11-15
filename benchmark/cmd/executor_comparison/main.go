package main

import (
	"context"
	"fmt"
	"time"

	"github.com/imran31415/godemode/pkg/executor"
)

func main() {
	// Simple test code
	testCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello from generated code!")
	sum := 0
	for i := 1; i <= 100; i++ {
		sum += i
	}
	fmt.Printf("Sum: %d\n", sum)
}
`

	fmt.Println("====================================================================================================")
	fmt.Println("EXECUTOR COMPARISON: WASM vs Interpreter")
	fmt.Println("====================================================================================================\n")

	// Test WASM Executor (with compilation overhead)
	fmt.Println("Testing WASM Executor (TinyGo compilation + WASM runtime)...")
	wasmExec := executor.NewExecutor()

	wasmStart := time.Now()
	wasmResult, wasmErr := wasmExec.Execute(context.Background(), testCode, 30*time.Second)
	wasmDuration := time.Since(wasmStart)

	fmt.Printf("WASM Executor Results:\n")
	fmt.Printf("  Duration: %v\n", wasmDuration)
	fmt.Printf("  Success: %v\n", wasmResult.Success)
	if wasmErr != nil {
		fmt.Printf("  Error: %v\n", wasmErr)
	} else {
		fmt.Printf("  Output:\n%s\n", wasmResult.Stdout)
	}

	// Test Interpreter Executor (no compilation!)
	fmt.Println("\n\nTesting Interpreter Executor (direct Go interpretation)...")
	interpExec := executor.NewInterpreterExecutor()

	interpStart := time.Now()
	interpResult, interpErr := interpExec.Execute(context.Background(), testCode, 30*time.Second)
	interpDuration := time.Since(interpStart)

	fmt.Printf("Interpreter Executor Results:\n")
	fmt.Printf("  Duration: %v\n", interpDuration)
	fmt.Printf("  Success: %v\n", interpResult.Success)
	if interpErr != nil {
		fmt.Printf("  Error: %v\n", interpErr)
	} else {
		fmt.Printf("  Output:\n%s\n", interpResult.Stdout)
	}

	// Comparison
	fmt.Println("\n\n====================================================================================================")
	fmt.Println("PERFORMANCE COMPARISON")
	fmt.Println("====================================================================================================")

	speedup := float64(wasmDuration) / float64(interpDuration)
	fmt.Printf("WASM Executor:        %v\n", wasmDuration)
	fmt.Printf("Interpreter Executor: %v\n", interpDuration)
	fmt.Printf("Speedup:              %.2fx faster\n", speedup)
	fmt.Printf("Time Saved:           %v\n", wasmDuration - interpDuration)

	if wasmResult.Success && interpResult.Success {
		fmt.Println("\n✓ Both executors completed successfully")
	}

	// Run multiple iterations to show consistent savings
	fmt.Println("\n\n====================================================================================================")
	fmt.Println("RUNNING 5 ITERATIONS FOR AVERAGE PERFORMANCE")
	fmt.Println("====================================================================================================\n")

	var wasmTotal, interpTotal time.Duration

	for i := 1; i <= 5; i++ {
		// WASM
		start := time.Now()
		wasmExec.Execute(context.Background(), testCode, 30*time.Second)
		wasmTime := time.Since(start)
		wasmTotal += wasmTime

		// Interpreter
		start = time.Now()
		interpExec.Execute(context.Background(), testCode, 30*time.Second)
		interpTime := time.Since(start)
		interpTotal += interpTime

		fmt.Printf("Iteration %d: WASM=%v, Interpreter=%v (%.2fx faster)\n",
			i, wasmTime, interpTime, float64(wasmTime)/float64(interpTime))
	}

	wasmAvg := wasmTotal / 5
	interpAvg := interpTotal / 5
	avgSpeedup := float64(wasmAvg) / float64(interpAvg)

	fmt.Println("\n====================================================================================================")
	fmt.Println("AVERAGE RESULTS (5 iterations)")
	fmt.Println("====================================================================================================")
	fmt.Printf("WASM Average:        %v\n", wasmAvg)
	fmt.Printf("Interpreter Average: %v\n", interpAvg)
	fmt.Printf("Average Speedup:     %.2fx faster\n", avgSpeedup)
	fmt.Printf("Average Time Saved:  %v per execution\n", wasmAvg - interpAvg)
}
