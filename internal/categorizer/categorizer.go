package categorizer

import (
	"fmt"
	"sync"

	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/pkg/config"
)

// Categorizer is the main API for transaction categorization
type Categorizer struct {
	config  *config.Config
	matcher *PatternMatcher
	loader  *Loader

	// mu protects patterns during concurrent access
	mu sync.RWMutex

	// patterns stores all loaded patterns
	patterns []*Pattern
}

// New creates a new Categorizer with the given configuration
func New(cfg *config.Config) (*Categorizer, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	loader := NewLoader()

	c := &Categorizer{
		config:   cfg,
		loader:   loader,
		patterns: make([]*Pattern, 0),
	}

	// Load patterns if file is configured
	if cfg.Files.PatternsFile != "" {
		if err := c.LoadPatterns(cfg.Files.PatternsFile); err != nil {
			// Don't fail if patterns file doesn't exist - allow categorizer to work without patterns
			// Other errors should be returned
			if !isNotExist(err) {
				return nil, fmt.Errorf("failed to load patterns: %w", err)
			}
		}
	}

	return c, nil
}

// LoadPatterns loads categorization patterns from a YAML file
func (c *Categorizer) LoadPatterns(path string) error {
	patterns, err := c.loader.LoadFile(path)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.patterns = patterns

	// Create new matcher with loaded patterns
	matcherConfig := MatcherConfig{
		EarlyExitThreshold: c.config.Categorization.ConfidenceThreshold,
		MaxAlternatives:    3,
	}
	c.matcher = NewPatternMatcherWithConfig(patterns, matcherConfig)

	return nil
}

// ReloadPatterns reloads patterns from the configured patterns file
func (c *Categorizer) ReloadPatterns() error {
	if c.config.Files.PatternsFile == "" {
		return fmt.Errorf("no patterns file configured")
	}
	return c.LoadPatterns(c.config.Files.PatternsFile)
}

// Suggest returns categorization suggestions for a transaction
func (c *Categorizer) Suggest(tx *beancount.Transaction) (*Suggestion, error) {
	if !c.config.Categorization.Enabled {
		return nil, nil
	}

	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.matcher == nil {
		return nil, nil
	}

	suggestion, err := c.matcher.Match(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to match pattern: %w", err)
	}

	return suggestion, nil
}

// SuggestAll returns all matching suggestions for a transaction
func (c *Categorizer) SuggestAll(tx *beancount.Transaction) ([]*Suggestion, error) {
	if !c.config.Categorization.Enabled {
		return nil, nil
	}

	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.matcher == nil {
		return nil, nil
	}

	suggestions, err := c.matcher.MatchAll(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to match patterns: %w", err)
	}

	return suggestions, nil
}

// Feedback records user feedback on a suggestion for learning
// If accepted is true, the pattern's statistics are updated positively
// If accepted is false, the pattern's statistics are updated negatively
func (c *Categorizer) Feedback(suggestion *Suggestion, accepted bool) error {
	if suggestion == nil {
		return fmt.Errorf("suggestion cannot be nil")
	}

	if suggestion.Pattern == nil {
		// No pattern to update (e.g., ML suggestion)
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.matcher == nil {
		return fmt.Errorf("no matcher initialized")
	}

	// Update pattern statistics
	if err := c.matcher.UpdateStatistics(suggestion.Pattern.ID, accepted); err != nil {
		return fmt.Errorf("failed to update statistics: %w", err)
	}

	// If configured to learn from edits, save updated patterns
	if c.config.Categorization.LearnFromEdits && c.config.Files.PatternsFile != "" {
		if err := c.savePatternsUnlocked(c.config.Files.PatternsFile); err != nil {
			// Log error but don't fail - statistics are already updated in memory
			return fmt.Errorf("failed to save patterns: %w", err)
		}
	}

	return nil
}

// SavePatterns saves all patterns to a YAML file
func (c *Categorizer) SavePatterns(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.savePatternsUnlocked(path)
}

// savePatternsUnlocked saves patterns without locking (internal use)
func (c *Categorizer) savePatternsUnlocked(path string) error {
	return c.loader.SaveFile(path, c.patterns)
}

// AddPattern adds a new pattern to the categorizer
func (c *Categorizer) AddPattern(pattern *Pattern) error {
	if pattern == nil {
		return fmt.Errorf("pattern cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for duplicate ID
	for _, p := range c.patterns {
		if p.ID == pattern.ID {
			return fmt.Errorf("pattern with ID '%s' already exists", pattern.ID)
		}
	}

	c.patterns = append(c.patterns, pattern)

	// Update matcher
	if c.matcher != nil {
		c.matcher.AddPattern(pattern)
	}

	return nil
}

// RemovePattern removes a pattern by ID
func (c *Categorizer) RemovePattern(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find and remove from patterns slice
	found := false
	for i, p := range c.patterns {
		if p.ID == id {
			c.patterns = append(c.patterns[:i], c.patterns[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("pattern not found: %s", id)
	}

	// Update matcher
	if c.matcher != nil {
		if !c.matcher.RemovePattern(id) {
			return fmt.Errorf("pattern not found in matcher: %s", id)
		}
	}

	return nil
}

// GetPattern returns a pattern by ID
func (c *Categorizer) GetPattern(id string) (*Pattern, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, p := range c.patterns {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, fmt.Errorf("pattern not found: %s", id)
}

// GetPatterns returns all patterns
func (c *Categorizer) GetPatterns() []*Pattern {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	patterns := make([]*Pattern, len(c.patterns))
	copy(patterns, c.patterns)
	return patterns
}

// PatternCount returns the number of loaded patterns
func (c *Categorizer) PatternCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.patterns)
}

// IsEnabled returns whether categorization is enabled
func (c *Categorizer) IsEnabled() bool {
	return c.config.Categorization.Enabled
}

// SetEnabled enables or disables categorization
func (c *Categorizer) SetEnabled(enabled bool) {
	c.config.Categorization.Enabled = enabled
}

// GetConfig returns the categorization configuration
func (c *Categorizer) GetConfig() config.CategorizationConfig {
	return c.config.Categorization
}

// isNotExist checks if an error is a "file not found" error
func isNotExist(err error) bool {
	if err == nil {
		return false
	}
	// Check error message for "patterns file not found" text
	errStr := err.Error()
	return len(errStr) >= 24 && errStr[:24] == "patterns file not found:"
}
