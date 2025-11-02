package categorizer

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// PatternFile represents the structure of a YAML patterns file
type PatternFile struct {
	// Version is the file format version (for future compatibility)
	Version string `yaml:"version"`

	// Patterns is the list of categorization patterns
	Patterns []PatternYAML `yaml:"patterns"`
}

// PatternYAML represents a pattern as stored in YAML
// This is separate from Pattern to allow for cleaner YAML structure
type PatternYAML struct {
	ID         string            `yaml:"id"`
	Name       string            `yaml:"name"`
	Pattern    string            `yaml:"pattern"`
	Category   string            `yaml:"category"`
	Fields     []string          `yaml:"fields,omitempty"`
	Priority   int               `yaml:"priority,omitempty"`
	Confidence float64           `yaml:"confidence,omitempty"`
	MinAmount  *float64          `yaml:"min_amount,omitempty"`
	MaxAmount  *float64          `yaml:"max_amount,omitempty"`
	Tags       []string          `yaml:"tags,omitempty"`
	Metadata   map[string]string `yaml:"metadata,omitempty"`
}

// LoaderConfig holds configuration for the pattern loader
type LoaderConfig struct {
	// DefaultConfidence is the confidence to use when not specified in YAML
	DefaultConfidence float64

	// DefaultFields is the fields to search when not specified in YAML
	DefaultFields []string

	// StrictMode enables strict validation (fail on any validation error)
	StrictMode bool
}

// DefaultLoaderConfig returns the default loader configuration
func DefaultLoaderConfig() LoaderConfig {
	return LoaderConfig{
		DefaultConfidence: 0.7,
		DefaultFields:     []string{"any"},
		StrictMode:        true,
	}
}

// Loader handles loading patterns from YAML files
type Loader struct {
	config LoaderConfig
}

// NewLoader creates a new pattern loader with default configuration
func NewLoader() *Loader {
	return NewLoaderWithConfig(DefaultLoaderConfig())
}

// NewLoaderWithConfig creates a new pattern loader with custom configuration
func NewLoaderWithConfig(config LoaderConfig) *Loader {
	return &Loader{
		config: config,
	}
}

// LoadFile loads patterns from a YAML file
func (l *Loader) LoadFile(path string) ([]*Pattern, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("patterns file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read patterns file: %w", err)
	}

	return l.LoadYAML(data)
}

// LoadYAML loads patterns from YAML data
func (l *Loader) LoadYAML(data []byte) ([]*Pattern, error) {
	var patternFile PatternFile

	// Parse YAML
	if err := yaml.Unmarshal(data, &patternFile); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate version (if specified)
	if patternFile.Version != "" && patternFile.Version != "1" {
		return nil, fmt.Errorf("unsupported patterns file version: %s (expected: 1)", patternFile.Version)
	}

	// Convert YAML patterns to Pattern structs
	patterns := make([]*Pattern, 0, len(patternFile.Patterns))
	errors := make([]error, 0)

	for i, yamlPattern := range patternFile.Patterns {
		pattern, err := l.convertPattern(yamlPattern, i)
		if err != nil {
			if l.config.StrictMode {
				return nil, fmt.Errorf("error in pattern %d (%s): %w", i, yamlPattern.ID, err)
			}
			errors = append(errors, fmt.Errorf("skipping pattern %d (%s): %w", i, yamlPattern.ID, err))
			continue
		}
		patterns = append(patterns, pattern)
	}

	// In non-strict mode, return patterns even if some had errors
	if len(errors) > 0 && !l.config.StrictMode {
		// Log errors but don't fail
		// In a real application, you might want to use a logger here
		for _, err := range errors {
			// For now, we'll just collect them
			_ = err
		}
	}

	return patterns, nil
}

// convertPattern converts a PatternYAML to a Pattern with validation
func (l *Loader) convertPattern(y PatternYAML, index int) (*Pattern, error) {
	// Validate required fields
	if y.ID == "" {
		return nil, fmt.Errorf("missing required field: id")
	}
	if y.Name == "" {
		return nil, fmt.Errorf("missing required field: name")
	}
	if y.Pattern == "" {
		return nil, fmt.Errorf("missing required field: pattern")
	}
	if y.Category == "" {
		return nil, fmt.Errorf("missing required field: category")
	}

	// Compile regex pattern
	regex, err := regexp.Compile(y.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Apply defaults
	fields := y.Fields
	if len(fields) == 0 {
		fields = l.config.DefaultFields
	}

	confidence := y.Confidence
	if confidence == 0 {
		confidence = l.config.DefaultConfidence
	}

	// Validate confidence range
	if confidence < 0 || confidence > 1 {
		return nil, fmt.Errorf("confidence must be between 0 and 1, got: %f", confidence)
	}

	// Validate fields
	for _, field := range fields {
		if field != "payee" && field != "narration" && field != "any" {
			return nil, fmt.Errorf("invalid field: %s (must be: payee, narration, or any)", field)
		}
	}

	// Validate amount constraints
	if y.MinAmount != nil && y.MaxAmount != nil && *y.MinAmount > *y.MaxAmount {
		return nil, fmt.Errorf("min_amount (%f) cannot be greater than max_amount (%f)", *y.MinAmount, *y.MaxAmount)
	}

	// Create pattern
	now := time.Now()
	pattern := &Pattern{
		ID:         y.ID,
		Name:       y.Name,
		Pattern:    y.Pattern,
		Regex:      regex,
		Category:   y.Category,
		Fields:     fields,
		Priority:   y.Priority,
		Confidence: confidence,
		MinAmount:  y.MinAmount,
		MaxAmount:  y.MaxAmount,
		Tags:       y.Tags,
		Metadata:   y.Metadata,
		Statistics: PatternStatistics{},
		Created:    now,
		Updated:    now,
	}

	return pattern, nil
}

// SaveFile saves patterns to a YAML file
func (l *Loader) SaveFile(path string, patterns []*Pattern) error {
	// Convert patterns to YAML structure
	yamlPatterns := make([]PatternYAML, len(patterns))
	for i, pattern := range patterns {
		yamlPatterns[i] = PatternYAML{
			ID:         pattern.ID,
			Name:       pattern.Name,
			Pattern:    pattern.Pattern,
			Category:   pattern.Category,
			Fields:     pattern.Fields,
			Priority:   pattern.Priority,
			Confidence: pattern.Confidence,
			MinAmount:  pattern.MinAmount,
			MaxAmount:  pattern.MaxAmount,
			Tags:       pattern.Tags,
			Metadata:   pattern.Metadata,
		}
	}

	patternFile := PatternFile{
		Version:  "1",
		Patterns: yamlPatterns,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&patternFile)
	if err != nil {
		return fmt.Errorf("failed to marshal patterns to YAML: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write patterns file: %w", err)
	}

	return nil
}

// ValidatePattern validates a single pattern without compiling it into a Pattern struct
func (l *Loader) ValidatePattern(y PatternYAML) error {
	_, err := l.convertPattern(y, 0)
	return err
}
