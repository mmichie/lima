package beancount

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Regular expressions for parsing
var (
	// Date format: YYYY-MM-DD
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

	// Transaction line: DATE FLAG ["PAYEE"] "NARRATION" [TAGS] [LINKS]
	// Examples:
	//   2025-01-01 * "Payee" "Narration"
	//   2025-01-01 ! "Narration"
	transactionRegex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\s+([*!])\s+(?:"([^"]*?)"\s+)?"([^"]*?)"(.*)$`)

	// Posting line: ACCOUNT [AMOUNT] [COMMODITY] [COST] [PRICE]
	// Must start with whitespace
	postingRegex = regexp.MustCompile(`^\s+([A-Z][A-Za-z0-9:_-]*)\s*(.*)$`)

	// Amount: NUMBER COMMODITY
	amountRegex = regexp.MustCompile(`^(-?\d+(?:\.\d+)?)\s+([A-Z][A-Z0-9._'-]{0,22}[A-Z0-9])`)

	// Metadata: KEY: VALUE
	metadataRegex = regexp.MustCompile(`^\s+([a-z][a-z0-9_-]*?):\s+(.+)$`)

	// Tag: #tag
	tagRegex = regexp.MustCompile(`#([A-Za-z0-9_-]+)`)

	// Link: ^link
	linkRegex = regexp.MustCompile(`\^([A-Za-z0-9_-]+)`)

	// Account name pattern
	accountRegex = regexp.MustCompile(`\b([A-Z][A-Za-z0-9]*(?::[A-Z][A-Za-z0-9]*)+)\b`)

	// Commodity pattern (2-24 uppercase letters/numbers)
	commodityRegex = regexp.MustCompile(`\b([A-Z][A-Z0-9._'-]{0,22}[A-Z0-9])\b`)
)

// parseTransactionIndexLine parses just enough to build an index entry
// Returns nil if line is not a transaction start
func parseTransactionIndexLine(line string, position int64, lineNumber int) *TransactionIndex {
	matches := transactionRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	date, err := time.Parse("2006-01-02", matches[1])
	if err != nil {
		return nil
	}

	payee := matches[3]
	if payee == "" {
		payee = matches[4] // If no payee, use narration
	}

	return &TransactionIndex{
		Date:         date,
		Payee:        payee,
		FilePosition: position,
		LineNumber:   lineNumber,
	}
}

// parseTransaction parses a complete transaction from the current scanner position
func parseTransaction(scanner *bufio.Scanner, startLine int) (*Transaction, error) {
	// Read first line (transaction header)
	if !scanner.Scan() {
		return nil, fmt.Errorf("unexpected end of file")
	}

	line := scanner.Text()
	matches := transactionRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("line %d: invalid transaction format: %s", startLine, line)
	}

	// Parse header
	date, err := time.Parse("2006-01-02", matches[1])
	if err != nil {
		return nil, fmt.Errorf("line %d: invalid date: %w", startLine, err)
	}

	tx := &Transaction{
		Date:      date,
		Flag:      matches[2],
		Payee:     matches[3],
		Narration: matches[4],
		Postings:  make([]Posting, 0),
		Metadata:  make(map[string]string),
		LineNumber: startLine,
	}

	// Extract tags and links from the rest of the line
	rest := matches[5]
	tx.Tags = extractTags(rest)
	tx.Links = extractLinks(rest)

	// Read postings and metadata until we hit a non-indented line or EOF
	lineNum := startLine
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Empty line or comment - continue
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), ";") {
			continue
		}

		// Non-indented line - end of transaction
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			break
		}

		// Try to parse as metadata
		if metaMatches := metadataRegex.FindStringSubmatch(line); metaMatches != nil {
			tx.Metadata[metaMatches[1]] = strings.Trim(metaMatches[2], `"`)
			continue
		}

		// Try to parse as posting
		if posting, err := parsePosting(line, lineNum); err == nil {
			tx.Postings = append(tx.Postings, *posting)
			continue
		}

		// Indented but not recognized - might be posting metadata
		// For now, skip (we'll enhance this later)
	}

	return tx, nil
}

// parsePosting parses a single posting line
func parsePosting(line string, lineNum int) (*Posting, error) {
	matches := postingRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("line %d: invalid posting format", lineNum)
	}

	posting := &Posting{
		Account:  matches[1],
		Metadata: make(map[string]string),
	}

	// Parse amount if present
	rest := strings.TrimSpace(matches[2])
	if rest != "" {
		amount, remaining, err := parseAmount(rest)
		if err == nil {
			posting.Amount = amount

			// Check for cost or price in remaining text
			// Format: @ price or @@ total_price or { cost }
			// For now, we'll handle simple @ price
			if strings.Contains(remaining, "@") {
				// Simple price extraction (enhance later)
				parts := strings.Split(remaining, "@")
				if len(parts) > 1 {
					priceAmount, _, err := parseAmount(strings.TrimSpace(parts[1]))
					if err == nil {
						posting.Price = priceAmount
					}
				}
			}
		}
	}

	return posting, nil
}

// parseAmount parses an amount from a string, returns amount and remaining text
func parseAmount(s string) (*Amount, string, error) {
	matches := amountRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil, s, fmt.Errorf("invalid amount format: %s", s)
	}

	number, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, s, fmt.Errorf("invalid number: %w", err)
	}

	amount := &Amount{
		Number:    number,
		Commodity: matches[2],
	}

	// Return remaining text after the amount
	remaining := s[len(matches[0]):]
	return amount, remaining, nil
}

// extractTags extracts all tags from a string
func extractTags(s string) []string {
	matches := tagRegex.FindAllStringSubmatch(s, -1)
	tags := make([]string, 0, len(matches))
	for _, match := range matches {
		tags = append(tags, match[1])
	}
	return tags
}

// extractLinks extracts all links from a string
func extractLinks(s string) []string {
	matches := linkRegex.FindAllStringSubmatch(s, -1)
	links := make([]string, 0, len(matches))
	for _, match := range matches {
		links = append(links, match[1])
	}
	return links
}

// extractAccountsAndCommodities extracts accounts and commodities from a line
// Used for building the index
func extractAccountsAndCommodities(line string) ([]string, []string) {
	accounts := make([]string, 0)
	commodities := make([]string, 0)

	// Extract accounts
	accountMatches := accountRegex.FindAllStringSubmatch(line, -1)
	for _, match := range accountMatches {
		accounts = append(accounts, match[1])
	}

	// Extract commodities
	commodityMatches := commodityRegex.FindAllStringSubmatch(line, -1)
	for _, match := range commodityMatches {
		// Filter out account names that might match commodity pattern
		isCommodity := true
		for _, acc := range accounts {
			if strings.Contains(acc, match[1]) {
				isCommodity = false
				break
			}
		}
		if isCommodity && len(match[1]) <= 10 { // Commodities are typically short
			commodities = append(commodities, match[1])
		}
	}

	return accounts, commodities
}
