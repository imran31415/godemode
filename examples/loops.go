package main

import "fmt"

func main() {
	// Demonstrate that loops work in WASM
	// This is a key advantage of code mode over function calling

	// Example 1: Creating multiple items in a loop
	fmt.Println("Example 1: Processing items in a loop")
	items := []string{"apple", "banana", "cherry", "date", "elderberry"}

	for i, item := range items {
		fmt.Printf("Item %d: %s (length: %d)\n", i+1, item, len(item))
	}

	// Example 2: Conditional logic with accumulation
	fmt.Println("\nExample 2: Filtering and accumulating")
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evenSum := 0
	oddSum := 0

	for _, num := range numbers {
		if num%2 == 0 {
			evenSum += num
		} else {
			oddSum += num
		}
	}

	fmt.Printf("Even sum: %d, Odd sum: %d\n", evenSum, oddSum)

	// Example 3: Nested loops
	fmt.Println("\nExample 3: Multiplication table")
	for i := 1; i <= 5; i++ {
		for j := 1; j <= 5; j++ {
			fmt.Printf("%3d ", i*j)
		}
		fmt.Println()
	}
}
