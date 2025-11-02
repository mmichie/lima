package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/pkg/config"
)

func TestUIModelCreation(t *testing.T) {
	// Create a temporary test file
	content := `2025-01-01 * "Test" "Transaction"
  Assets:Checking  -100.00 USD
  Expenses:Test  100.00 USD
`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	// Open file
	file, err := beancount.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create UI model with default config
	cfg := config.DefaultConfig()
	model := New(file, cfg)

	// Verify initial state
	if model.currentView != DashboardView {
		t.Errorf("expected initial view to be Dashboard, got %d", model.currentView)
	}

	if model.file == nil {
		t.Error("expected file to be set")
	}

	// Verify views are initialized (by checking they render without error)
	model.width = 80
	model.height = 24
	model.ready = true
	model.dashboard = model.dashboard.SetSize(80, 20)

	dashboardView := model.dashboard.View()
	if dashboardView == "" {
		t.Error("dashboard not initialized properly")
	}
}

func TestUIRendering(t *testing.T) {
	// Create a temporary test file
	content := `2025-01-01 * "Test" "Transaction"
  Assets:Checking  -100.00 USD
  Expenses:Test  100.00 USD
`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	// Open file
	file, err := beancount.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create UI model with default config
	cfg := config.DefaultConfig()
	model := New(file, cfg)

	// Set window size to trigger ready state
	model.width = 80
	model.height = 24
	model.ready = true

	// Update view sizes
	model.dashboard = model.dashboard.SetSize(80, 20)
	model.transactions = model.transactions.SetSize(80, 20)
	model.accounts = model.accounts.SetSize(80, 20)

	// Test rendering each view
	views := []ViewType{DashboardView, TransactionsView, AccountsView}
	for _, view := range views {
		model.currentView = view
		output := model.View()

		if output == "" {
			t.Errorf("view %d produced empty output", view)
		}

		if output == "Loading..." {
			t.Errorf("view %d still loading", view)
		}
	}
}

func TestKeyboardNavigation(t *testing.T) {
	// Create a temporary test file
	content := `2025-01-01 * "Test" "Transaction"
  Assets:Checking  -100.00 USD
  Expenses:Test  100.00 USD
`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	// Open file
	file, err := beancount.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create UI model with default config
	cfg := config.DefaultConfig()
	model := New(file, cfg)
	model.width = 80
	model.height = 24
	model.ready = true

	// Test view switching with key bindings
	tests := []struct {
		key      string
		expected ViewType
	}{
		{"1", DashboardView},
		{"2", TransactionsView},
		{"3", AccountsView},
		{"4", ReportsView},
	}

	for _, tt := range tests {
		// Simulate key press would require more complex setup
		// For now, just verify the keys are configured
		switch tt.key {
		case "1":
			if model.keys.Dashboard.Keys()[0] != tt.key {
				t.Errorf("dashboard key not configured correctly")
			}
		case "2":
			if model.keys.Transactions.Keys()[0] != tt.key {
				t.Errorf("transactions key not configured correctly")
			}
		case "3":
			if model.keys.Accounts.Keys()[0] != tt.key {
				t.Errorf("accounts key not configured correctly")
			}
		case "4":
			if model.keys.Reports.Keys()[0] != tt.key {
				t.Errorf("reports key not configured correctly")
			}
		}
	}
}

func createTempFile(t *testing.T, content string) string {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_ui_beancount.txt")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return tmpFile
}
