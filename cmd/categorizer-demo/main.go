package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/internal/categorizer"
	"github.com/mmichie/lima/pkg/config"
)

func main() {
	fmt.Println("=== Lima Categorization Engine Demo ===\n")

	// Create a config with the patterns file
	cfg := config.DefaultConfig()
	cfg.Files.PatternsFile = "examples/patterns.yaml"
	cfg.Categorization.Enabled = true
	cfg.Categorization.ConfidenceThreshold = 0.8

	// Create the categorizer
	fmt.Println("Loading categorizer...")
	cat, err := categorizer.New(cfg)
	if err != nil {
		fmt.Printf("Error creating categorizer: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d patterns\n\n", cat.PatternCount())

	// Test transactions
	testTransactions := []struct {
		payee     string
		narration string
	}{
		{"STARBUCKS #12345", "Morning coffee"},
		{"SAFEWAY #789", "Weekly groceries"},
		{"UBER *TRIP", "Ride to airport"},
		{"AMAZON.COM", "Online shopping"},
		{"NETFLIX.COM", "Monthly subscription"},
		{"Comcast", "Internet bill"},
		{"Unknown Merchant", "Mystery purchase"},
		{"Local Coffee Shop", "Afternoon coffee"},
		{"WHOLE FOODS MARKET", "Organic groceries"},
		{"SHELL OIL", "Gas for car"},
	}

	fmt.Println("Testing categorization suggestions:\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, tt := range testTransactions {
		tx := &beancount.Transaction{
			Date:      time.Now(),
			Payee:     tt.payee,
			Narration: tt.narration,
		}

		fmt.Printf("\n%d. Transaction: %s - %s\n", i+1, tt.payee, tt.narration)

		// Get best suggestion
		suggestion, err := cat.Suggest(tx)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			continue
		}

		if suggestion == nil {
			fmt.Printf("   ❌ No matches found\n")
		} else {
			// Show the suggestion with confidence
			confidencePercent := suggestion.Confidence * 100
			emoji := "✅"
			if confidencePercent < 80 {
				emoji = "⚠️"
			}

			fmt.Printf("   %s Suggestion: %s (%.0f%% confidence)\n",
				emoji, suggestion.Category, confidencePercent)
			fmt.Printf("      Pattern: %s\n", suggestion.Pattern.Name)
			fmt.Printf("      Reason: %s\n", suggestion.Reason)

			// Show alternatives if any
			if len(suggestion.Alternatives) > 0 {
				fmt.Printf("      Alternatives:\n")
				for _, alt := range suggestion.Alternatives {
					fmt.Printf("        - %s (%.0f%%)\n", alt.Category, alt.Confidence*100)
				}
			}
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Demonstrate feedback
	fmt.Println("\n=== Testing Feedback System ===\n")

	tx := &beancount.Transaction{
		Date:      time.Now(),
		Payee:     "STARBUCKS #999",
		Narration: "Coffee",
	}

	suggestion, _ := cat.Suggest(tx)
	if suggestion != nil {
		fmt.Printf("Before feedback - Pattern statistics:\n")
		pattern, _ := cat.GetPattern(suggestion.Pattern.ID)
		fmt.Printf("  Match count: %d\n", pattern.Statistics.MatchCount)
		fmt.Printf("  Accept count: %d\n", pattern.Statistics.AcceptCount)
		fmt.Printf("  Accuracy: %.0f%%\n\n", pattern.Statistics.Accuracy*100)

		// Provide positive feedback
		fmt.Println("Providing positive feedback...")
		cat.Feedback(suggestion, true)

		pattern, _ = cat.GetPattern(suggestion.Pattern.ID)
		fmt.Printf("\nAfter feedback - Pattern statistics:\n")
		fmt.Printf("  Match count: %d\n", pattern.Statistics.MatchCount)
		fmt.Printf("  Accept count: %d\n", pattern.Statistics.AcceptCount)
		fmt.Printf("  Accuracy: %.0f%%\n", pattern.Statistics.Accuracy*100)
	}

	// Show all patterns with statistics
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n=== Pattern Statistics Summary ===\n")

	patterns := cat.GetPatterns()
	matchedPatterns := 0
	for _, p := range patterns {
		if p.Statistics.MatchCount > 0 {
			matchedPatterns++
			fmt.Printf("%-30s Matches: %d, Accuracy: %.0f%%, Priority: %d\n",
				p.Name,
				p.Statistics.MatchCount,
				p.Statistics.Accuracy*100,
				p.Priority)
		}
	}

	fmt.Printf("\nTotal patterns: %d\n", len(patterns))
	fmt.Printf("Patterns used: %d\n", matchedPatterns)

	fmt.Println("\n✨ Demo complete! ✨")
}
