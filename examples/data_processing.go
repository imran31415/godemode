package main

import (
	"fmt"
	"strings"
)

type Transaction struct {
	ID       int
	Type     string
	Amount   float64
	Category string
}

func main() {
	// Simulate what an LLM might generate for a multi-step task
	// This demonstrates the power of code mode: complex logic in one pass

	fmt.Println("=== Multi-Step Data Processing Example ===\n")

	// Step 1: Create sample data
	transactions := []Transaction{
		{1, "income", 5000.00, "salary"},
		{2, "expense", 1200.00, "rent"},
		{3, "expense", 150.00, "utilities"},
		{4, "income", 500.00, "freelance"},
		{5, "expense", 75.00, "groceries"},
		{6, "expense", 200.00, "entertainment"},
		{7, "income", 100.00, "interest"},
	}

	fmt.Printf("Processing %d transactions...\n\n", len(transactions))

	// Step 2: Calculate totals by type
	var totalIncome, totalExpense float64
	incomeCount, expenseCount := 0, 0

	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Amount
			incomeCount++
		} else {
			totalExpense += tx.Amount
			expenseCount++
		}
	}

	// Step 3: Group by category
	categories := make(map[string]float64)
	for _, tx := range transactions {
		categories[tx.Category] += tx.Amount
	}

	// Step 4: Print summary report
	fmt.Println("Financial Summary:")
	fmt.Printf("  Total Income:  $%.2f (%d transactions)\n", totalIncome, incomeCount)
	fmt.Printf("  Total Expense: $%.2f (%d transactions)\n", totalExpense, expenseCount)
	fmt.Printf("  Net:           $%.2f\n", totalIncome-totalExpense)

	fmt.Println("\nBy Category:")
	for category, amount := range categories {
		fmt.Printf("  %-15s $%.2f\n", strings.Title(category)+":", amount)
	}

	// Step 5: Find largest transactions
	fmt.Println("\nTop 3 Transactions:")
	// Simple bubble sort for demo
	sorted := make([]Transaction, len(transactions))
	copy(sorted, transactions)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Amount < sorted[j+1].Amount {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	for i := 0; i < 3 && i < len(sorted); i++ {
		tx := sorted[i]
		fmt.Printf("  %d. %s - $%.2f (%s)\n", i+1, strings.Title(tx.Category), tx.Amount, tx.Type)
	}

	fmt.Println("\n✓ Processing complete!")
}
