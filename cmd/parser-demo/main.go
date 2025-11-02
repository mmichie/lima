package main

import (
	"fmt"
	"os"

	"github.com/mmichie/lima/internal/beancount"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: parser-demo <beancount-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Open the file
	fmt.Printf("Opening %s...\n\n", filename)
	file, err := beancount.Open(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Print summary
	fmt.Printf("✓ Successfully parsed Beancount file\n")
	fmt.Printf("  Transactions: %d\n", file.TransactionCount())
	fmt.Printf("  Accounts: %d\n", len(file.GetAccounts()))
	fmt.Printf("  Commodities: %d\n\n", len(file.GetCommodities()))

	// Print accounts
	fmt.Println("Accounts:")
	for _, account := range file.GetAccounts() {
		fmt.Printf("  - %s\n", account)
	}
	fmt.Println()

	// Print first 5 transactions
	fmt.Println("Recent Transactions:")
	count := file.TransactionCount()
	if count > 5 {
		count = 5
	}

	for i := 0; i < count; i++ {
		tx, err := file.GetTransaction(i)
		if err != nil {
			fmt.Printf("Error getting transaction %d: %v\n", i, err)
			continue
		}

		fmt.Printf("\n%s %s \"%s\" \"%s\"\n",
			tx.Date.Format("2006-01-02"),
			tx.Flag,
			tx.Payee,
			tx.Narration)

		if len(tx.Tags) > 0 {
			fmt.Printf("  Tags: %v\n", tx.Tags)
		}

		if len(tx.Links) > 0 {
			fmt.Printf("  Links: %v\n", tx.Links)
		}

		for _, posting := range tx.Postings {
			if posting.Amount != nil {
				fmt.Printf("  %s  %s %s\n",
					posting.Account,
					posting.Amount.Number.StringFixed(2),
					posting.Amount.Commodity)
			} else {
				fmt.Printf("  %s\n", posting.Account)
			}
		}
	}
}
