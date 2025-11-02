package beancount

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// File represents an opened Beancount file with lazy loading support
type File struct {
	path  string
	file  *os.File
	index *Index
	cache *Cache
}

// Index stores positions of all directives in the file for lazy loading
type Index struct {
	transactions []TransactionIndex
	accounts     []string
	commodities  []string
}

// TransactionIndex stores metadata about a transaction for quick access
type TransactionIndex struct {
	Date         time.Time
	Payee        string
	FilePath     string // Path to the file containing this transaction
	FilePosition int64  // Position within that file
	LineNumber   int
}

// Cache stores recently accessed transactions
type Cache struct {
	transactions map[int]*Transaction
	maxSize      int
}

// Open opens a Beancount file and builds an index
func Open(path string) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	f := &File{
		path: path,
		file: file,
		cache: &Cache{
			transactions: make(map[int]*Transaction),
			maxSize:      100, // Cache last 100 transactions
		},
	}

	// Build index on first open
	if err := f.buildIndex(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to build index: %w", err)
	}

	return f, nil
}

// Close closes the Beancount file
func (f *File) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

// TransactionCount returns the total number of transactions in the file
func (f *File) TransactionCount() int {
	return len(f.index.transactions)
}

// GetTransaction retrieves a transaction by index (0-based)
// Uses lazy loading - only parses the transaction when requested
func (f *File) GetTransaction(index int) (*Transaction, error) {
	if index < 0 || index >= len(f.index.transactions) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}

	// Check cache first
	if tx, ok := f.cache.transactions[index]; ok {
		return tx, nil
	}

	// Not in cache - load from file
	txIndex := f.index.transactions[index]
	tx, err := f.parseTransactionAt(txIndex.FilePath, txIndex.FilePosition, txIndex.LineNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction at index %d: %w", index, err)
	}

	// Add to cache
	f.cache.transactions[index] = tx

	// Evict oldest if cache is full (simple FIFO for now)
	if len(f.cache.transactions) > f.cache.maxSize {
		// Find smallest index and remove it
		minIdx := index
		for idx := range f.cache.transactions {
			if idx < minIdx {
				minIdx = idx
			}
		}
		delete(f.cache.transactions, minIdx)
	}

	return tx, nil
}

// GetTransactionsByDateRange returns all transactions within a date range
func (f *File) GetTransactionsByDateRange(start, end time.Time) ([]*Transaction, error) {
	var transactions []*Transaction

	for i, txIndex := range f.index.transactions {
		if (txIndex.Date.Equal(start) || txIndex.Date.After(start)) &&
			(txIndex.Date.Equal(end) || txIndex.Date.Before(end)) {
			tx, err := f.GetTransaction(i)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, tx)
		}
	}

	return transactions, nil
}

// GetAccounts returns all unique account names found in the file
func (f *File) GetAccounts() []string {
	return f.index.accounts
}

// GetCommodities returns all unique commodities found in the file
func (f *File) GetCommodities() []string {
	return f.index.commodities
}

// buildIndex scans the entire file and builds an index of all directives
func (f *File) buildIndex() error {
	f.index = &Index{
		transactions: make([]TransactionIndex, 0),
		accounts:     make([]string, 0),
		commodities:  make([]string, 0),
	}

	accountSet := make(map[string]bool)
	commoditySet := make(map[string]bool)
	includedFiles := make(map[string]bool)

	// Process the main file and all includes recursively
	if err := f.processFile(f.path, accountSet, commoditySet, includedFiles); err != nil {
		return err
	}

	return nil
}

// processFile recursively processes a file and all its includes
func (f *File) processFile(filePath string, accountSet, commoditySet map[string]bool, includedFiles map[string]bool) error {
	// Check if already included to avoid infinite loops
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}

	if includedFiles[absPath] {
		return nil // Already processed
	}
	includedFiles[absPath] = true

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max line size

	var position int64
	lineNumber := 0
	baseDir := filepath.Dir(filePath)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check for include directive
		if matches := includeRegex.FindStringSubmatch(line); matches != nil {
			includePath := matches[1]
			// Resolve relative paths
			if !filepath.IsAbs(includePath) {
				includePath = filepath.Join(baseDir, includePath)
			}
			// Recursively process included file
			if err := f.processFile(includePath, accountSet, commoditySet, includedFiles); err != nil {
				return fmt.Errorf("error processing include %s: %w", includePath, err)
			}
			position += int64(len(scanner.Bytes()) + 1)
			continue
		}

		// Try to parse as transaction start
		if txIndex := parseTransactionIndexLine(line, filePath, position, lineNumber); txIndex != nil {
			f.index.transactions = append(f.index.transactions, *txIndex)
		}

		// Extract accounts and commodities
		accounts, commodities := extractAccountsAndCommodities(line)
		for _, acc := range accounts {
			if !accountSet[acc] {
				accountSet[acc] = true
				f.index.accounts = append(f.index.accounts, acc)
			}
		}
		for _, comm := range commodities {
			if !commoditySet[comm] {
				commoditySet[comm] = true
				f.index.commodities = append(f.index.commodities, comm)
			}
		}

		position += int64(len(scanner.Bytes()) + 1) // +1 for newline
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file %s: %w", filePath, err)
	}

	return nil
}

// parseTransactionAt seeks to a position and parses a complete transaction
func (f *File) parseTransactionAt(filePath string, position int64, lineNumber int) (*Transaction, error) {
	// Open the correct file (might be an included file, not the main file)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Seek to the position
	if _, err := file.Seek(position, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to position %d: %w", position, err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// Parse the transaction starting at this position
	tx, err := parseTransaction(scanner, lineNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction at line %d: %w", lineNumber, err)
	}

	tx.FilePosition = position
	tx.LineNumber = lineNumber

	return tx, nil
}
