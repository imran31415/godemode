package main

import "fmt"

func main() {
	// Simple Hello World example
	fmt.Println("Hello from GoDeMode!")

	// Demonstrate basic computation
	result := fibonacci(10)
	fmt.Printf("Fibonacci(10) = %d\n", result)

	// Show that loops work
	sum := 0
	for i := 1; i <= 10; i++ {
		sum += i
	}
	fmt.Printf("Sum of 1-10 = %d\n", sum)
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
