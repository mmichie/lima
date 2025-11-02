package beancount

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestParseBasicTransaction(t *testing.T) {
	// Create a temporary test file
	content := `2025-01-01 * "Coffee Shop" "Morning coffee"
  Assets:Checking  -5.00 USD
  Expenses:Food:DiningOut  5.00 USD

2025-01-02 * "Grocery Store" "Weekly groceries"
  Assets:Checking  -150.00 USD
  Expenses:Food:Groceries  150.00 USD
`

	tmpFile, err := createTempFile(content)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Open and parse
	f, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	// Check transaction count
	count := f.TransactionCount()
	if count != 2 {
		t.Errorf("expected 2 transactions, got %d", count)
	}

	// Check first transaction
	tx, err := f.GetTransaction(0)
	if err != nil {
		t.Fatalf("failed to get transaction 0: %v", err)
	}

	expectedDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !tx.Date.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, tx.Date)
	}

	if tx.Flag != "*" {
		t.Errorf("expected flag '*', got '%s'", tx.Flag)
	}

	if tx.Payee != "Coffee Shop" {
		t.Errorf("expected payee 'Coffee Shop', got '%s'", tx.Payee)
	}

	if tx.Narration != "Morning coffee" {
		t.Errorf("expected narration 'Morning coffee', got '%s'", tx.Narration)
	}

	if len(tx.Postings) != 2 {
		t.Fatalf("expected 2 postings, got %d", len(tx.Postings))
	}

	// Check first posting
	posting := tx.Postings[0]
	if posting.Account != "Assets:Checking" {
		t.Errorf("expected account 'Assets:Checking', got '%s'", posting.Account)
	}

	if posting.Amount == nil {
		t.Fatal("expected amount, got nil")
	}

	expectedAmount := decimal.NewFromFloat(-5.00)
	if !posting.Amount.Number.Equal(expectedAmount) {
		t.Errorf("expected amount %v, got %v", expectedAmount, posting.Amount.Number)
	}

	if posting.Amount.Commodity != "USD" {
		t.Errorf("expected commodity 'USD', got '%s'", posting.Amount.Commodity)
	}
}

func TestParseTransactionWithTags(t *testing.T) {
	content := `2025-01-01 * "Payee" "Narration" #tag1 #tag2
  Assets:Checking  -100.00 USD
  Expenses:Test  100.00 USD
`

	tmpFile, err := createTempFile(content)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	f, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	tx, err := f.GetTransaction(0)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}

	if len(tx.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tx.Tags))
	}

	if tx.Tags[0] != "tag1" || tx.Tags[1] != "tag2" {
		t.Errorf("unexpected tags: %v", tx.Tags)
	}
}

func TestParseTransactionWithLinks(t *testing.T) {
	content := `2025-01-01 * "Payee" "Narration" ^link1 ^link2
  Assets:Checking  -100.00 USD
  Expenses:Test  100.00 USD
`

	tmpFile, err := createTempFile(content)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	f, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	tx, err := f.GetTransaction(0)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}

	if len(tx.Links) != 2 {
		t.Errorf("expected 2 links, got %d", len(tx.Links))
	}

	if tx.Links[0] != "link1" || tx.Links[1] != "link2" {
		t.Errorf("unexpected links: %v", tx.Links)
	}
}

func TestGetTransactionsByDateRange(t *testing.T) {
	content := `2025-01-01 * "Store" "Item 1"
  Assets:Checking  -10.00 USD
  Expenses:Test  10.00 USD

2025-01-05 * "Store" "Item 2"
  Assets:Checking  -20.00 USD
  Expenses:Test  20.00 USD

2025-01-10 * "Store" "Item 3"
  Assets:Checking  -30.00 USD
  Expenses:Test  30.00 USD
`

	tmpFile, err := createTempFile(content)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	f, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	// Get transactions from Jan 2 to Jan 8
	start := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)

	txs, err := f.GetTransactionsByDateRange(start, end)
	if err != nil {
		t.Fatalf("failed to get transactions by date range: %v", err)
	}

	// Should only get the Jan 5 transaction
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}

	if len(txs) > 0 && txs[0].Narration != "Item 2" {
		t.Errorf("expected 'Item 2', got '%s'", txs[0].Narration)
	}
}

func TestGetAccounts(t *testing.T) {
	content := `2025-01-01 * "Test" "Test"
  Assets:Checking  -100.00 USD
  Expenses:Food:Groceries  100.00 USD
`

	tmpFile, err := createTempFile(content)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	f, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	accounts := f.GetAccounts()
	if len(accounts) < 2 {
		t.Errorf("expected at least 2 accounts, got %d", len(accounts))
	}

	// Check that both accounts are present
	hasChecking := false
	hasGroceries := false
	for _, acc := range accounts {
		if acc == "Assets:Checking" {
			hasChecking = true
		}
		if acc == "Expenses:Food:Groceries" {
			hasGroceries = true
		}
	}

	if !hasChecking {
		t.Error("expected to find Assets:Checking")
	}
	if !hasGroceries {
		t.Error("expected to find Expenses:Food:Groceries")
	}
}

func TestCache(t *testing.T) {
	// Create file with multiple transactions
	var content string
	for i := 1; i <= 150; i++ {
		content += "2025-01-01 * \"Store\" \"Transaction\"\n"
		content += "  Assets:Checking  -10.00 USD\n"
		content += "  Expenses:Test  10.00 USD\n\n"
	}

	tmpFile, err := createTempFile(content)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	f, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	// Access transaction 0 - should be cached
	_, err = f.GetTransaction(0)
	if err != nil {
		t.Fatalf("failed to get transaction 0: %v", err)
	}

	// Access transaction 149 - should evict transaction 0 from cache
	_, err = f.GetTransaction(149)
	if err != nil {
		t.Fatalf("failed to get transaction 149: %v", err)
	}

	// Access transaction 0 again - should still work (re-parse from file)
	_, err = f.GetTransaction(0)
	if err != nil {
		t.Fatalf("failed to get transaction 0 again: %v", err)
	}
}

// Helper function to create a temporary test file
func createTempFile(content string) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_beancount.txt")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return tmpFile, nil
}
