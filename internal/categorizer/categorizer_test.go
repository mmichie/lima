package categorizer

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/pkg/config"
)

func TestNew_DefaultConfig(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if c == nil {
		t.Fatal("Expected categorizer, got nil")
	}

	if c.config == nil {
		t.Error("Expected config to be set")
	}
}

func TestNew_WithConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Categorization.Enabled = true
	cfg.Categorization.ConfidenceThreshold = 0.95

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !c.IsEnabled() {
		t.Error("Expected categorization to be enabled")
	}
}

func TestNew_LoadPatternsFromConfig(t *testing.T) {
	// Create temp patterns file
	tmpDir := t.TempDir()
	patternsFile := filepath.Join(tmpDir, "patterns.yaml")

	yaml := `
version: "1"
patterns:
  - id: test
    name: Test Pattern
    pattern: "TEST"
    category: Expenses:Test
`
	if err := os.WriteFile(patternsFile, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to create patterns file: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Files.PatternsFile = patternsFile

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if c.PatternCount() != 1 {
		t.Errorf("Expected 1 pattern, got %d", c.PatternCount())
	}
}

func TestNew_MissingPatternsFile(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Files.PatternsFile = "/nonexistent/patterns.yaml"

	// Should not error - missing patterns file is OK
	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Should not error on missing patterns file: %v", err)
	}

	if c.PatternCount() != 0 {
		t.Errorf("Expected 0 patterns, got %d", c.PatternCount())
	}
}

func TestCategorizer_LoadPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	patternsFile := filepath.Join(tmpDir, "patterns.yaml")

	yaml := `
patterns:
  - id: starbucks
    name: Starbucks
    pattern: "STARBUCKS"
    category: Expenses:Food:DiningOut
    confidence: 0.9
  - id: safeway
    name: Safeway
    pattern: "SAFEWAY"
    category: Expenses:Food:Groceries
    confidence: 0.85
`
	if err := os.WriteFile(patternsFile, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to create patterns file: %v", err)
	}

	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	if err := c.LoadPatterns(patternsFile); err != nil {
		t.Fatalf("Failed to load patterns: %v", err)
	}

	if c.PatternCount() != 2 {
		t.Errorf("Expected 2 patterns, got %d", c.PatternCount())
	}
}

func TestCategorizer_ReloadPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	patternsFile := filepath.Join(tmpDir, "patterns.yaml")

	// Initial patterns
	yaml1 := `
patterns:
  - id: test1
    name: Test 1
    pattern: "TEST1"
    category: Expenses:Test1
`
	if err := os.WriteFile(patternsFile, []byte(yaml1), 0644); err != nil {
		t.Fatalf("Failed to create patterns file: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Files.PatternsFile = patternsFile

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}
	if c.PatternCount() != 1 {
		t.Errorf("Expected 1 pattern initially, got %d", c.PatternCount())
	}

	// Update patterns file
	yaml2 := `
patterns:
  - id: test1
    name: Test 1
    pattern: "TEST1"
    category: Expenses:Test1
  - id: test2
    name: Test 2
    pattern: "TEST2"
    category: Expenses:Test2
`
	if err := os.WriteFile(patternsFile, []byte(yaml2), 0644); err != nil {
		t.Fatalf("Failed to update patterns file: %v", err)
	}

	// Reload
	if err := c.ReloadPatterns(); err != nil {
		t.Fatalf("Failed to reload patterns: %v", err)
	}

	if c.PatternCount() != 2 {
		t.Errorf("Expected 2 patterns after reload, got %d", c.PatternCount())
	}
}

func TestCategorizer_Suggest(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	// Add a pattern
	pattern := &Pattern{
		ID:         "test",
		Name:       "Test",
		Pattern:    "STARBUCKS",
		Regex:      regexp.MustCompile("STARBUCKS"),
		Category:   "Expenses:Food:DiningOut",
		Confidence: 0.9,
		Fields:     []string{"payee"},
	}
	c.AddPattern(pattern)

	// Create matcher
	c.matcher = NewPatternMatcher([]*Pattern{pattern})

	tx := &beancount.Transaction{
		Payee:     "STARBUCKS #12345",
		Narration: "Coffee",
	}

	suggestion, err := c.Suggest(tx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if suggestion == nil {
		t.Fatal("Expected suggestion, got nil")
	}

	if suggestion.Category != "Expenses:Food:DiningOut" {
		t.Errorf("Expected category 'Expenses:Food:DiningOut', got '%s'", suggestion.Category)
	}
}

func TestCategorizer_Suggest_Disabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Categorization.Enabled = false

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	tx := &beancount.Transaction{
		Payee: "TEST",
	}

	suggestion, err := c.Suggest(tx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if suggestion != nil {
		t.Error("Expected nil suggestion when disabled")
	}
}

func TestCategorizer_Suggest_NilTransaction(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	_, err = c.Suggest(nil)
	if err == nil {
		t.Error("Expected error for nil transaction")
	}
}

func TestCategorizer_SuggestAll(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	// Add multiple patterns
	pattern1 := &Pattern{
		ID:         "starbucks",
		Name:       "Starbucks",
		Pattern:    "STARBUCKS",
		Regex:      regexp.MustCompile("STARBUCKS"),
		Category:   "Expenses:Food:DiningOut",
		Confidence: 0.9,
		Fields:     []string{"payee"},
	}

	pattern2 := &Pattern{
		ID:         "coffee",
		Name:       "Coffee",
		Pattern:    "(?i)coffee",
		Regex:      regexp.MustCompile("(?i)coffee"),
		Category:   "Expenses:Food:Coffee",
		Confidence: 0.7,
		Fields:     []string{"narration"},
	}

	c.AddPattern(pattern1)
	c.AddPattern(pattern2)

	// Create matcher
	c.matcher = NewPatternMatcher([]*Pattern{pattern1, pattern2})

	tx := &beancount.Transaction{
		Payee:     "STARBUCKS #12345",
		Narration: "Morning coffee",
	}

	suggestions, err := c.SuggestAll(tx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(suggestions) != 2 {
		t.Fatalf("Expected 2 suggestions, got %d", len(suggestions))
	}
}

func TestCategorizer_Feedback(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Categorization.LearnFromEdits = false // Disable learning for this test

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern := &Pattern{
		ID:         "test",
		Name:       "Test",
		Pattern:    "TEST",
		Regex:      regexp.MustCompile("TEST"),
		Category:   "Expenses:Test",
		Confidence: 0.8,
		Fields:     []string{"payee"},
	}
	c.AddPattern(pattern)
	c.matcher = NewPatternMatcher([]*Pattern{pattern})

	tx := &beancount.Transaction{
		Payee: "TEST",
	}

	suggestion, _ := c.Suggest(tx)

	// Provide positive feedback
	err = c.Feedback(suggestion, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that statistics were updated
	p, _ := c.GetPattern("test")
	if p.Statistics.AcceptCount != 1 {
		t.Errorf("Expected AcceptCount 1, got %d", p.Statistics.AcceptCount)
	}
}

func TestCategorizer_Feedback_NilSuggestion(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	err = c.Feedback(nil, true)
	if err == nil {
		t.Error("Expected error for nil suggestion")
	}
}

func TestCategorizer_Feedback_NoPattern(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	suggestion := &Suggestion{
		Category:   "Expenses:Test",
		Confidence: 0.8,
		Pattern:    nil, // No pattern
		Source:     SourceML,
	}

	// Should not error when pattern is nil (e.g., ML suggestions)
	err = c.Feedback(suggestion, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCategorizer_Feedback_LearnFromEdits(t *testing.T) {
	tmpDir := t.TempDir()
	patternsFile := filepath.Join(tmpDir, "patterns.yaml")

	cfg := config.DefaultConfig()
	cfg.Files.PatternsFile = patternsFile
	cfg.Categorization.LearnFromEdits = true

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern := &Pattern{
		ID:         "test",
		Name:       "Test",
		Pattern:    "TEST",
		Regex:      regexp.MustCompile("TEST"),
		Category:   "Expenses:Test",
		Confidence: 0.8,
		Fields:     []string{"payee"},
	}
	c.AddPattern(pattern)
	c.matcher = NewPatternMatcher([]*Pattern{pattern})

	tx := &beancount.Transaction{
		Payee: "TEST",
	}

	suggestion, _ := c.Suggest(tx)

	// Provide feedback - should save patterns
	err = c.Feedback(suggestion, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(patternsFile); os.IsNotExist(err) {
		t.Error("Expected patterns file to be created")
	}
}

func TestCategorizer_AddPattern(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern := &Pattern{
		ID:       "test",
		Name:     "Test",
		Pattern:  "TEST",
		Category: "Expenses:Test",
	}

	err = c.AddPattern(pattern)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if c.PatternCount() != 1 {
		t.Errorf("Expected 1 pattern, got %d", c.PatternCount())
	}
}

func TestCategorizer_AddPattern_Duplicate(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern1 := &Pattern{
		ID:       "test",
		Name:     "Test 1",
		Pattern:  "TEST1",
		Category: "Expenses:Test1",
	}

	pattern2 := &Pattern{
		ID:       "test", // Same ID
		Name:     "Test 2",
		Pattern:  "TEST2",
		Category: "Expenses:Test2",
	}

	c.AddPattern(pattern1)

	err = c.AddPattern(pattern2)
	if err == nil {
		t.Error("Expected error for duplicate ID")
	}
}

func TestCategorizer_RemovePattern(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern := &Pattern{
		ID:       "test",
		Name:     "Test",
		Pattern:  "TEST",
		Category: "Expenses:Test",
	}
	c.AddPattern(pattern)
	c.matcher = NewPatternMatcher([]*Pattern{pattern})

	err = c.RemovePattern("test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if c.PatternCount() != 0 {
		t.Errorf("Expected 0 patterns, got %d", c.PatternCount())
	}
}

func TestCategorizer_RemovePattern_NotFound(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	err = c.RemovePattern("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent pattern")
	}
}

func TestCategorizer_GetPattern(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern := &Pattern{
		ID:       "test",
		Name:     "Test Pattern",
		Pattern:  "TEST",
		Category: "Expenses:Test",
	}
	c.AddPattern(pattern)

	p, err := c.GetPattern("test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.Name != "Test Pattern" {
		t.Errorf("Expected name 'Test Pattern', got '%s'", p.Name)
	}
}

func TestCategorizer_GetPattern_NotFound(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	_, err = c.GetPattern("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent pattern")
	}
}

func TestCategorizer_GetPatterns(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern1 := &Pattern{ID: "test1", Name: "Test 1", Pattern: "TEST1", Category: "Expenses:Test1"}
	pattern2 := &Pattern{ID: "test2", Name: "Test 2", Pattern: "TEST2", Category: "Expenses:Test2"}

	c.AddPattern(pattern1)
	c.AddPattern(pattern2)

	patterns := c.GetPatterns()
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect internal state)
	patterns[0] = nil
	if c.PatternCount() != 2 {
		t.Error("GetPatterns should return a copy, not the internal slice")
	}
}

func TestCategorizer_SavePatterns(t *testing.T) {
	tmpDir := t.TempDir()
	patternsFile := filepath.Join(tmpDir, "patterns.yaml")

	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	pattern := &Pattern{
		ID:         "test",
		Name:       "Test",
		Pattern:    "TEST",
		Category:   "Expenses:Test",
		Confidence: 0.8,
	}
	c.AddPattern(pattern)

	err = c.SavePatterns(patternsFile)
	if err != nil {
		t.Fatalf("Failed to save patterns: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(patternsFile); os.IsNotExist(err) {
		t.Error("Expected patterns file to be created")
	}

	// Load and verify
	c2, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}
	c2.LoadPatterns(patternsFile)

	if c2.PatternCount() != 1 {
		t.Errorf("Expected 1 pattern after loading, got %d", c2.PatternCount())
	}
}

func TestCategorizer_IsEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Categorization.Enabled = true

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	if !c.IsEnabled() {
		t.Error("Expected categorization to be enabled")
	}
}

func TestCategorizer_SetEnabled(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	c.SetEnabled(false)
	if c.IsEnabled() {
		t.Error("Expected categorization to be disabled")
	}

	c.SetEnabled(true)
	if !c.IsEnabled() {
		t.Error("Expected categorization to be enabled")
	}
}

func TestCategorizer_GetConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Categorization.ConfidenceThreshold = 0.95
	cfg.Categorization.LearnFromEdits = true

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create categorizer: %v", err)
	}

	categCfg := c.GetConfig()
	if categCfg.ConfidenceThreshold != 0.95 {
		t.Errorf("Expected confidence threshold 0.95, got %f", categCfg.ConfidenceThreshold)
	}
	if !categCfg.LearnFromEdits {
		t.Error("Expected LearnFromEdits to be true")
	}
}
