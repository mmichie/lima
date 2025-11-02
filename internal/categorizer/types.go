package categorizer

import (
	"regexp"
	"time"

	"github.com/mmichie/lima/internal/beancount"
)

// Pattern represents a categorization pattern for matching transactions
type Pattern struct {
	// ID is a unique identifier for this pattern
	ID string

	// Name is a human-readable name for this pattern
	Name string

	// Pattern is the regex pattern to match against transaction fields
	Pattern string

	// Regex is the compiled regular expression (cached)
	Regex *regexp.Regexp

	// Category is the account to suggest when this pattern matches
	Category string

	// Fields specifies which transaction fields to match against
	// Valid values: "payee", "narration", "any"
	Fields []string

	// Priority determines the order of pattern matching (higher = checked first)
	// Used for conflict resolution when multiple patterns match
	Priority int

	// Confidence is the base confidence score for this pattern (0.0 to 1.0)
	// Higher values indicate more reliable patterns
	Confidence float64

	// MinAmount is the minimum transaction amount for this pattern to apply
	MinAmount *float64

	// MaxAmount is the maximum transaction amount for this pattern to apply
	MaxAmount *float64

	// Tags are optional tags that must be present on the transaction
	Tags []string

	// Metadata stores additional pattern-specific data
	Metadata map[string]string

	// Statistics track pattern usage and accuracy
	Statistics PatternStatistics

	// Created is when this pattern was created
	Created time.Time

	// Updated is when this pattern was last modified
	Updated time.Time
}

// PatternStatistics tracks usage and performance metrics for a pattern
type PatternStatistics struct {
	// MatchCount is the total number of times this pattern has matched
	MatchCount int

	// AcceptCount is how many times suggestions from this pattern were accepted
	AcceptCount int

	// RejectCount is how many times suggestions were rejected
	RejectCount int

	// LastMatched is the most recent time this pattern matched
	LastMatched time.Time

	// Accuracy is the acceptance rate (AcceptCount / (AcceptCount + RejectCount))
	Accuracy float64
}

// Suggestion represents a categorization suggestion for a transaction
type Suggestion struct {
	// Transaction is the transaction being categorized
	Transaction *beancount.Transaction

	// Category is the suggested account
	Category string

	// Confidence is the confidence score for this suggestion (0.0 to 1.0)
	Confidence float64

	// Pattern is the pattern that generated this suggestion (may be nil for ML suggestions)
	Pattern *Pattern

	// Source indicates where this suggestion came from
	// Valid values: "pattern", "ml", "history", "manual"
	Source SuggestionSource

	// Reason is a human-readable explanation for this suggestion
	Reason string

	// Alternatives are other possible suggestions with lower confidence
	Alternatives []Alternative

	// Metadata stores additional suggestion-specific data
	Metadata map[string]string

	// Created is when this suggestion was generated
	Created time.Time
}

// SuggestionSource indicates the origin of a categorization suggestion
type SuggestionSource string

const (
	// SourcePattern indicates the suggestion came from a pattern match
	SourcePattern SuggestionSource = "pattern"

	// SourceML indicates the suggestion came from machine learning
	SourceML SuggestionSource = "ml"

	// SourceHistory indicates the suggestion came from transaction history
	SourceHistory SuggestionSource = "history"

	// SourceManual indicates the suggestion was manually created
	SourceManual SuggestionSource = "manual"
)

// Alternative represents an alternative categorization suggestion
type Alternative struct {
	// Category is the alternative account
	Category string

	// Confidence is the confidence score for this alternative
	Confidence float64

	// Reason is why this alternative is suggested
	Reason string
}

// Matches checks if this pattern matches the given transaction
func (p *Pattern) Matches(tx *beancount.Transaction) bool {
	if p.Regex == nil {
		return false
	}

	// Check amount constraints
	if p.MinAmount != nil || p.MaxAmount != nil {
		if !p.matchesAmount(tx) {
			return false
		}
	}

	// Check tag requirements
	if len(p.Tags) > 0 && !p.matchesTags(tx) {
		return false
	}

	// Check field matches
	if len(p.Fields) == 0 || contains(p.Fields, "any") {
		// Match against any field
		if p.Regex.MatchString(tx.Payee) || p.Regex.MatchString(tx.Narration) {
			return true
		}
	} else {
		// Match against specific fields
		for _, field := range p.Fields {
			switch field {
			case "payee":
				if p.Regex.MatchString(tx.Payee) {
					return true
				}
			case "narration":
				if p.Regex.MatchString(tx.Narration) {
					return true
				}
			}
		}
	}

	return false
}

// matchesAmount checks if the transaction amount falls within the pattern's constraints
func (p *Pattern) matchesAmount(tx *beancount.Transaction) bool {
	// Find the largest posting amount (typically the expense)
	var maxAmount float64
	for _, posting := range tx.Postings {
		if posting.Amount != nil {
			amount, _ := posting.Amount.Number.Float64()
			if amount > maxAmount {
				maxAmount = amount
			}
			if amount < 0 {
				maxAmount = -amount
			}
		}
	}

	if p.MinAmount != nil && maxAmount < *p.MinAmount {
		return false
	}

	if p.MaxAmount != nil && maxAmount > *p.MaxAmount {
		return false
	}

	return true
}

// matchesTags checks if the transaction has all required tags
func (p *Pattern) matchesTags(tx *beancount.Transaction) bool {
	txTags := make(map[string]bool)
	for _, tag := range tx.Tags {
		txTags[tag] = true
	}

	for _, requiredTag := range p.Tags {
		if !txTags[requiredTag] {
			return false
		}
	}

	return true
}

// UpdateStatistics updates pattern statistics after a match
func (p *Pattern) UpdateStatistics(accepted bool) {
	p.Statistics.MatchCount++
	if accepted {
		p.Statistics.AcceptCount++
	} else {
		p.Statistics.RejectCount++
	}
	p.Statistics.LastMatched = time.Now()

	// Recalculate accuracy
	total := p.Statistics.AcceptCount + p.Statistics.RejectCount
	if total > 0 {
		p.Statistics.Accuracy = float64(p.Statistics.AcceptCount) / float64(total)
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
