package categorizer

import (
	"fmt"
	"sort"
	"time"

	"github.com/mmichie/lima/internal/beancount"
)

// PatternMatcher matches transactions against a collection of patterns
type PatternMatcher struct {
	patterns []*Pattern

	// EarlyExitThreshold is the confidence threshold for early exit optimization
	// When a pattern matches with confidence >= this threshold, matching stops
	// Set to 1.0 to disable early exit
	EarlyExitThreshold float64

	// MaxAlternatives is the maximum number of alternative suggestions to include
	MaxAlternatives int
}

// MatcherConfig holds configuration for a PatternMatcher
type MatcherConfig struct {
	// EarlyExitThreshold is the confidence threshold for early exit (default: 0.95)
	EarlyExitThreshold float64

	// MaxAlternatives is the maximum number of alternatives to return (default: 3)
	MaxAlternatives int
}

// DefaultMatcherConfig returns the default matcher configuration
func DefaultMatcherConfig() MatcherConfig {
	return MatcherConfig{
		EarlyExitThreshold: 0.95,
		MaxAlternatives:    3,
	}
}

// NewPatternMatcher creates a new pattern matcher with the given patterns
func NewPatternMatcher(patterns []*Pattern) *PatternMatcher {
	config := DefaultMatcherConfig()
	return NewPatternMatcherWithConfig(patterns, config)
}

// NewPatternMatcherWithConfig creates a new pattern matcher with custom configuration
func NewPatternMatcherWithConfig(patterns []*Pattern, config MatcherConfig) *PatternMatcher {
	// Sort patterns by priority (higher priority first)
	sortedPatterns := make([]*Pattern, len(patterns))
	copy(sortedPatterns, patterns)
	sort.Slice(sortedPatterns, func(i, j int) bool {
		// First sort by priority (descending)
		if sortedPatterns[i].Priority != sortedPatterns[j].Priority {
			return sortedPatterns[i].Priority > sortedPatterns[j].Priority
		}
		// Then by confidence (descending)
		if sortedPatterns[i].Confidence != sortedPatterns[j].Confidence {
			return sortedPatterns[i].Confidence > sortedPatterns[j].Confidence
		}
		// Finally by accuracy (descending)
		return sortedPatterns[i].Statistics.Accuracy > sortedPatterns[j].Statistics.Accuracy
	})

	return &PatternMatcher{
		patterns:           sortedPatterns,
		EarlyExitThreshold: config.EarlyExitThreshold,
		MaxAlternatives:    config.MaxAlternatives,
	}
}

// AddPattern adds a pattern to the matcher and re-sorts by priority
func (pm *PatternMatcher) AddPattern(pattern *Pattern) {
	pm.patterns = append(pm.patterns, pattern)
	pm.sortPatterns()
}

// RemovePattern removes a pattern by ID
func (pm *PatternMatcher) RemovePattern(id string) bool {
	for i, p := range pm.patterns {
		if p.ID == id {
			pm.patterns = append(pm.patterns[:i], pm.patterns[i+1:]...)
			return true
		}
	}
	return false
}

// GetPattern returns a pattern by ID
func (pm *PatternMatcher) GetPattern(id string) *Pattern {
	for _, p := range pm.patterns {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Match finds the best matching pattern for a transaction
// Uses early exit optimization - stops when a high-confidence match is found
func (pm *PatternMatcher) Match(tx *beancount.Transaction) (*Suggestion, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	var bestMatch *Pattern
	var allMatches []*Pattern

	// Iterate through patterns in priority order
	for _, pattern := range pm.patterns {
		if pattern.Matches(tx) {
			if bestMatch == nil {
				bestMatch = pattern
			}
			allMatches = append(allMatches, pattern)

			// Early exit if confidence is high enough
			if pattern.Confidence >= pm.EarlyExitThreshold {
				break
			}
		}
	}

	// No matches found
	if bestMatch == nil {
		return nil, nil
	}

	// Create suggestion from best match
	suggestion := pm.createSuggestion(tx, bestMatch, allMatches)

	return suggestion, nil
}

// MatchAll finds all matching patterns for a transaction (no early exit)
// Returns suggestions sorted by confidence
func (pm *PatternMatcher) MatchAll(tx *beancount.Transaction) ([]*Suggestion, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	var matches []*Pattern

	// Find all matching patterns
	for _, pattern := range pm.patterns {
		if pattern.Matches(tx) {
			matches = append(matches, pattern)
		}
	}

	// No matches found
	if len(matches) == 0 {
		return nil, nil
	}

	// Create suggestions
	suggestions := make([]*Suggestion, 0, len(matches))
	for _, pattern := range matches {
		// For MatchAll, we don't include alternatives in each suggestion
		suggestion := &Suggestion{
			Transaction: tx,
			Category:    pattern.Category,
			Confidence:  pm.calculateConfidence(pattern),
			Pattern:     pattern,
			Source:      SourcePattern,
			Reason:      pm.generateReason(pattern),
			Created:     time.Now(),
		}
		suggestions = append(suggestions, suggestion)
	}

	// Sort by confidence (highest first)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	return suggestions, nil
}

// createSuggestion creates a suggestion from a pattern match with alternatives
func (pm *PatternMatcher) createSuggestion(tx *beancount.Transaction, best *Pattern, allMatches []*Pattern) *Suggestion {
	suggestion := &Suggestion{
		Transaction: tx,
		Category:    best.Category,
		Confidence:  pm.calculateConfidence(best),
		Pattern:     best,
		Source:      SourcePattern,
		Reason:      pm.generateReason(best),
		Created:     time.Now(),
	}

	// Add alternatives (excluding the best match)
	alternatives := make([]Alternative, 0, pm.MaxAlternatives)
	for _, pattern := range allMatches {
		if pattern.ID != best.ID && len(alternatives) < pm.MaxAlternatives {
			alt := Alternative{
				Category:   pattern.Category,
				Confidence: pm.calculateConfidence(pattern),
				Reason:     pm.generateReason(pattern),
			}
			alternatives = append(alternatives, alt)
		}
	}
	suggestion.Alternatives = alternatives

	return suggestion
}

// calculateConfidence computes the final confidence score for a pattern
// Takes into account the pattern's base confidence and its historical accuracy
func (pm *PatternMatcher) calculateConfidence(pattern *Pattern) float64 {
	baseConfidence := pattern.Confidence

	// If we have historical data, blend it with base confidence
	if pattern.Statistics.AcceptCount+pattern.Statistics.RejectCount > 0 {
		accuracy := pattern.Statistics.Accuracy

		// Weight: 70% base confidence, 30% historical accuracy
		// This gives more weight to the configured confidence but adjusts based on reality
		return (baseConfidence * 0.7) + (accuracy * 0.3)
	}

	return baseConfidence
}

// generateReason creates a human-readable explanation for why a pattern matched
func (pm *PatternMatcher) generateReason(pattern *Pattern) string {
	reason := fmt.Sprintf("Matched pattern '%s'", pattern.Name)

	// Add accuracy info if available
	totalMatches := pattern.Statistics.AcceptCount + pattern.Statistics.RejectCount
	if totalMatches > 0 {
		reason += fmt.Sprintf(" (accuracy: %.0f%% from %d previous matches)",
			pattern.Statistics.Accuracy*100, totalMatches)
	}

	return reason
}

// sortPatterns sorts patterns by priority, confidence, and accuracy
func (pm *PatternMatcher) sortPatterns() {
	sort.Slice(pm.patterns, func(i, j int) bool {
		// First sort by priority (descending)
		if pm.patterns[i].Priority != pm.patterns[j].Priority {
			return pm.patterns[i].Priority > pm.patterns[j].Priority
		}
		// Then by confidence (descending)
		if pm.patterns[i].Confidence != pm.patterns[j].Confidence {
			return pm.patterns[i].Confidence > pm.patterns[j].Confidence
		}
		// Finally by accuracy (descending)
		return pm.patterns[i].Statistics.Accuracy > pm.patterns[j].Statistics.Accuracy
	})
}

// Len returns the number of patterns in the matcher
func (pm *PatternMatcher) Len() int {
	return len(pm.patterns)
}

// UpdateStatistics updates the statistics for a pattern after user feedback
func (pm *PatternMatcher) UpdateStatistics(patternID string, accepted bool) error {
	pattern := pm.GetPattern(patternID)
	if pattern == nil {
		return fmt.Errorf("pattern not found: %s", patternID)
	}

	pattern.UpdateStatistics(accepted)

	// Re-sort patterns since accuracy may have changed
	pm.sortPatterns()

	return nil
}
