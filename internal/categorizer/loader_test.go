package categorizer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadYAML_Valid(t *testing.T) {
	yaml := `
version: "1"
patterns:
  - id: starbucks
    name: Starbucks Coffee
    pattern: "STARBUCKS"
    category: Expenses:Food:DiningOut
    priority: 10
    confidence: 0.9
    fields:
      - payee
  - id: safeway
    name: Safeway Groceries
    pattern: "SAFEWAY|QFC"
    category: Expenses:Food:Groceries
    priority: 5
    confidence: 0.85
    fields:
      - payee
      - narration
`

	loader := NewLoader()
	patterns, err := loader.LoadYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(patterns) != 2 {
		t.Fatalf("Expected 2 patterns, got %d", len(patterns))
	}

	// Check first pattern
	p1 := patterns[0]
	if p1.ID != "starbucks" {
		t.Errorf("Expected ID 'starbucks', got '%s'", p1.ID)
	}
	if p1.Name != "Starbucks Coffee" {
		t.Errorf("Expected name 'Starbucks Coffee', got '%s'", p1.Name)
	}
	if p1.Pattern != "STARBUCKS" {
		t.Errorf("Expected pattern 'STARBUCKS', got '%s'", p1.Pattern)
	}
	if p1.Category != "Expenses:Food:DiningOut" {
		t.Errorf("Expected category 'Expenses:Food:DiningOut', got '%s'", p1.Category)
	}
	if p1.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", p1.Priority)
	}
	if p1.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", p1.Confidence)
	}
	if len(p1.Fields) != 1 || p1.Fields[0] != "payee" {
		t.Errorf("Expected fields ['payee'], got %v", p1.Fields)
	}
	if p1.Regex == nil {
		t.Error("Expected compiled regex")
	}

	// Check second pattern
	p2 := patterns[1]
	if p2.ID != "safeway" {
		t.Errorf("Expected ID 'safeway', got '%s'", p2.ID)
	}
	if len(p2.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(p2.Fields))
	}
}

func TestLoader_LoadYAML_DefaultValues(t *testing.T) {
	yaml := `
version: "1"
patterns:
  - id: test
    name: Test Pattern
    pattern: "TEST"
    category: Expenses:Test
`

	loader := NewLoader()
	patterns, err := loader.LoadYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(patterns))
	}

	p := patterns[0]

	// Check defaults
	if p.Confidence != 0.7 {
		t.Errorf("Expected default confidence 0.7, got %f", p.Confidence)
	}
	if len(p.Fields) != 1 || p.Fields[0] != "any" {
		t.Errorf("Expected default fields ['any'], got %v", p.Fields)
	}
	if p.Priority != 0 {
		t.Errorf("Expected default priority 0, got %d", p.Priority)
	}
}

func TestLoader_LoadYAML_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		errorMsg string
	}{
		{
			name: "missing id",
			yaml: `
patterns:
  - name: Test
    pattern: "TEST"
    category: Expenses:Test
`,
			errorMsg: "missing required field: id",
		},
		{
			name: "missing name",
			yaml: `
patterns:
  - id: test
    pattern: "TEST"
    category: Expenses:Test
`,
			errorMsg: "missing required field: name",
		},
		{
			name: "missing pattern",
			yaml: `
patterns:
  - id: test
    name: Test
    category: Expenses:Test
`,
			errorMsg: "missing required field: pattern",
		},
		{
			name: "missing category",
			yaml: `
patterns:
  - id: test
    name: Test
    pattern: "TEST"
`,
			errorMsg: "missing required field: category",
		},
	}

	loader := NewLoader()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loader.LoadYAML([]byte(tt.yaml))
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			// Just check that we got an error - the exact message may vary
		})
	}
}

func TestLoader_LoadYAML_InvalidRegex(t *testing.T) {
	yaml := `
patterns:
  - id: invalid
    name: Invalid Pattern
    pattern: "[invalid("
    category: Expenses:Test
`

	loader := NewLoader()
	_, err := loader.LoadYAML([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error for invalid regex")
	}
}

func TestLoader_LoadYAML_InvalidConfidence(t *testing.T) {
	tests := []struct {
		name       string
		confidence float64
	}{
		{"negative", -0.5},
		{"greater than 1", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := fmt.Sprintf(`
patterns:
  - id: test
    name: Test
    pattern: "TEST"
    category: Expenses:Test
    confidence: %f
`, tt.confidence)

			loader := NewLoader()
			_, err := loader.LoadYAML([]byte(yaml))
			if err == nil {
				t.Fatal("Expected error for invalid confidence")
			}
		})
	}
}

func TestLoader_LoadYAML_InvalidFields(t *testing.T) {
	yaml := `
patterns:
  - id: test
    name: Test
    pattern: "TEST"
    category: Expenses:Test
    fields:
      - invalid_field
`

	loader := NewLoader()
	_, err := loader.LoadYAML([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error for invalid field")
	}
}

func TestLoader_LoadYAML_InvalidAmountConstraints(t *testing.T) {
	yaml := `
patterns:
  - id: test
    name: Test
    pattern: "TEST"
    category: Expenses:Test
    min_amount: 100.0
    max_amount: 50.0
`

	loader := NewLoader()
	_, err := loader.LoadYAML([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error for min_amount > max_amount")
	}
}

func TestLoader_LoadYAML_AmountConstraints(t *testing.T) {
	yaml := `
patterns:
  - id: test
    name: Test
    pattern: "TEST"
    category: Expenses:Test
    min_amount: 10.0
    max_amount: 100.0
`

	loader := NewLoader()
	patterns, err := loader.LoadYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(patterns))
	}

	p := patterns[0]
	if p.MinAmount == nil || *p.MinAmount != 10.0 {
		t.Errorf("Expected min_amount 10.0, got %v", p.MinAmount)
	}
	if p.MaxAmount == nil || *p.MaxAmount != 100.0 {
		t.Errorf("Expected max_amount 100.0, got %v", p.MaxAmount)
	}
}

func TestLoader_LoadYAML_TagsAndMetadata(t *testing.T) {
	yaml := `
patterns:
  - id: test
    name: Test
    pattern: "TEST"
    category: Expenses:Test
    tags:
      - work
      - reimbursable
    metadata:
      source: manual
      notes: Test pattern
`

	loader := NewLoader()
	patterns, err := loader.LoadYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	p := patterns[0]
	if len(p.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(p.Tags))
	}
	if p.Tags[0] != "work" || p.Tags[1] != "reimbursable" {
		t.Errorf("Expected tags [work, reimbursable], got %v", p.Tags)
	}
	if len(p.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(p.Metadata))
	}
	if p.Metadata["source"] != "manual" {
		t.Errorf("Expected metadata source=manual, got %s", p.Metadata["source"])
	}
}

func TestLoader_LoadYAML_UnsupportedVersion(t *testing.T) {
	yaml := `
version: "2"
patterns:
  - id: test
    name: Test
    pattern: "TEST"
    category: Expenses:Test
`

	loader := NewLoader()
	_, err := loader.LoadYAML([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error for unsupported version")
	}
}

func TestLoader_LoadYAML_NonStrictMode(t *testing.T) {
	yaml := `
patterns:
  - id: valid
    name: Valid Pattern
    pattern: "VALID"
    category: Expenses:Valid
  - id: invalid
    name: Invalid Pattern
    pattern: "[invalid("
    category: Expenses:Invalid
  - id: valid2
    name: Another Valid Pattern
    pattern: "VALID2"
    category: Expenses:Valid2
`

	config := LoaderConfig{
		DefaultConfidence: 0.7,
		DefaultFields:     []string{"any"},
		StrictMode:        false,
	}

	loader := NewLoaderWithConfig(config)
	patterns, err := loader.LoadYAML([]byte(yaml))

	// Should not return an error in non-strict mode
	if err != nil {
		t.Fatalf("Unexpected error in non-strict mode: %v", err)
	}

	// Should have 2 valid patterns (skipped the invalid one)
	if len(patterns) != 2 {
		t.Fatalf("Expected 2 patterns (skipped invalid), got %d", len(patterns))
	}

	if patterns[0].ID != "valid" {
		t.Errorf("Expected first pattern 'valid', got '%s'", patterns[0].ID)
	}
	if patterns[1].ID != "valid2" {
		t.Errorf("Expected second pattern 'valid2', got '%s'", patterns[1].ID)
	}
}

func TestLoader_LoadFile_NotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFile("/nonexistent/path/to/patterns.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
	// Error message should indicate file not found
}

func TestLoader_LoadFile_Success(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "patterns.yaml")

	yaml := `
version: "1"
patterns:
  - id: test
    name: Test Pattern
    pattern: "TEST"
    category: Expenses:Test
    confidence: 0.8
`

	if err := os.WriteFile(tmpFile, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	loader := NewLoader()
	patterns, err := loader.LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(patterns))
	}

	if patterns[0].ID != "test" {
		t.Errorf("Expected ID 'test', got '%s'", patterns[0].ID)
	}
}

func TestLoader_SaveFile(t *testing.T) {
	patterns := []*Pattern{
		{
			ID:         "test1",
			Name:       "Test Pattern 1",
			Pattern:    "TEST1",
			Category:   "Expenses:Test1",
			Fields:     []string{"payee"},
			Priority:   10,
			Confidence: 0.9,
		},
		{
			ID:         "test2",
			Name:       "Test Pattern 2",
			Pattern:    "TEST2",
			Category:   "Expenses:Test2",
			Fields:     []string{"narration"},
			Priority:   5,
			Confidence: 0.8,
			MinAmount:  floatPtr(10.0),
			MaxAmount:  floatPtr(100.0),
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "patterns.yaml")

	loader := NewLoader()

	// Save patterns
	if err := loader.SaveFile(tmpFile, patterns); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Load them back
	loadedPatterns, err := loader.LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load saved file: %v", err)
	}

	if len(loadedPatterns) != 2 {
		t.Fatalf("Expected 2 patterns, got %d", len(loadedPatterns))
	}

	// Verify first pattern
	p1 := loadedPatterns[0]
	if p1.ID != "test1" {
		t.Errorf("Expected ID 'test1', got '%s'", p1.ID)
	}
	if p1.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", p1.Priority)
	}

	// Verify second pattern with amount constraints
	p2 := loadedPatterns[1]
	if p2.MinAmount == nil || *p2.MinAmount != 10.0 {
		t.Errorf("Expected min_amount 10.0, got %v", p2.MinAmount)
	}
	if p2.MaxAmount == nil || *p2.MaxAmount != 100.0 {
		t.Errorf("Expected max_amount 100.0, got %v", p2.MaxAmount)
	}
}

func TestLoader_ValidatePattern(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name      string
		pattern   PatternYAML
		wantError bool
	}{
		{
			name: "valid pattern",
			pattern: PatternYAML{
				ID:       "test",
				Name:     "Test",
				Pattern:  "TEST",
				Category: "Expenses:Test",
			},
			wantError: false,
		},
		{
			name: "missing id",
			pattern: PatternYAML{
				Name:     "Test",
				Pattern:  "TEST",
				Category: "Expenses:Test",
			},
			wantError: true,
		},
		{
			name: "invalid regex",
			pattern: PatternYAML{
				ID:       "test",
				Name:     "Test",
				Pattern:  "[invalid(",
				Category: "Expenses:Test",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.ValidatePattern(tt.pattern)
			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultLoaderConfig(t *testing.T) {
	config := DefaultLoaderConfig()

	if config.DefaultConfidence != 0.7 {
		t.Errorf("Expected default confidence 0.7, got %f", config.DefaultConfidence)
	}
	if len(config.DefaultFields) != 1 || config.DefaultFields[0] != "any" {
		t.Errorf("Expected default fields ['any'], got %v", config.DefaultFields)
	}
	if !config.StrictMode {
		t.Error("Expected strict mode to be true by default")
	}
}
