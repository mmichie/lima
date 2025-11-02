package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/internal/ui"
	"github.com/mmichie/lima/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadDefault()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check for file argument or use config default
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	} else if cfg.Files.DefaultLedger != "" {
		filename = cfg.Files.DefaultLedger
	} else {
		// Fallback to sample file for development
		filename = "testdata/sample.beancount"
	}

	// Open the Beancount file
	file, err := beancount.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create the TUI with config
	m := ui.New(file, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
