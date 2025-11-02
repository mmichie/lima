package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// File paths
	Files FilesConfig `yaml:"files"`

	// UI preferences
	UI UIConfig `yaml:"ui"`

	// Theme configuration
	Theme ThemeConfig `yaml:"theme"`

	// Keybindings
	Keybindings KeybindingsConfig `yaml:"keybindings"`

	// Categorization settings
	Categorization CategorizationConfig `yaml:"categorization"`
}

// FilesConfig contains file path settings
type FilesConfig struct {
	DefaultLedger string `yaml:"default_ledger"`
	PatternsFile  string `yaml:"patterns_file"`
}

// UIConfig contains UI preferences
type UIConfig struct {
	DefaultView    string `yaml:"default_view"`     // "dashboard", "transactions", "accounts", "reports"
	PageSize       int    `yaml:"page_size"`        // Number of items per page
	DateFormat     string `yaml:"date_format"`      // Date format string
	ShowLineNumbers bool   `yaml:"show_line_numbers"`
	CompactMode    bool   `yaml:"compact_mode"`
}

// ThemeConfig contains theme settings
type ThemeConfig struct {
	Primary    string `yaml:"primary"`    // Primary accent color
	Secondary  string `yaml:"secondary"`  // Secondary accent color
	Success    string `yaml:"success"`    // Success/positive color
	Warning    string `yaml:"warning"`    // Warning color
	Error      string `yaml:"error"`      // Error/negative color
	Muted      string `yaml:"muted"`      // Muted/secondary text color
	Text       string `yaml:"text"`       // Primary text color
	Background string `yaml:"background"` // Background color
}

// KeybindingsConfig contains keybinding settings
type KeybindingsConfig struct {
	Quit         []string `yaml:"quit"`
	Help         []string `yaml:"help"`
	Dashboard    []string `yaml:"dashboard"`
	Transactions []string `yaml:"transactions"`
	Accounts     []string `yaml:"accounts"`
	Reports      []string `yaml:"reports"`
	Up           []string `yaml:"up"`
	Down         []string `yaml:"down"`
	PageUp       []string `yaml:"page_up"`
	PageDown     []string `yaml:"page_down"`
	Top          []string `yaml:"top"`
	Bottom       []string `yaml:"bottom"`
	Select       []string `yaml:"select"`
	Back         []string `yaml:"back"`
}

// CategorizationConfig contains categorization settings
type CategorizationConfig struct {
	Enabled         bool    `yaml:"enabled"`
	AutoCategorize  bool    `yaml:"auto_categorize"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	LearnFromEdits  bool    `yaml:"learn_from_edits"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		Files: FilesConfig{
			DefaultLedger: filepath.Join(homeDir, "finances", "main.beancount"),
			PatternsFile:  filepath.Join(homeDir, ".config", "lima", "patterns.yaml"),
		},
		UI: UIConfig{
			DefaultView:    "dashboard",
			PageSize:       20,
			DateFormat:     "2006-01-02",
			ShowLineNumbers: false,
			CompactMode:    false,
		},
		Theme: ThemeConfig{
			Primary:    "#00D9FF",
			Secondary:  "#7D56F4",
			Success:    "#00FF00",
			Warning:    "#FFFF00",
			Error:      "#FF0000",
			Muted:      "#666666",
			Text:       "#FFFFFF",
			Background: "#1a1a1a",
		},
		Keybindings: KeybindingsConfig{
			Quit:         []string{"q", "ctrl+c"},
			Help:         []string{"?"},
			Dashboard:    []string{"1"},
			Transactions: []string{"2"},
			Accounts:     []string{"3"},
			Reports:      []string{"4"},
			Up:           []string{"up", "k"},
			Down:         []string{"down", "j"},
			PageUp:       []string{"pgup", "ctrl+b"},
			PageDown:     []string{"pgdown", "ctrl+f"},
			Top:          []string{"home", "g"},
			Bottom:       []string{"end", "G"},
			Select:       []string{"enter", "space"},
			Back:         []string{"esc", "backspace"},
		},
		Categorization: CategorizationConfig{
			Enabled:         true,
			AutoCategorize:  false,
			ConfidenceThreshold: 0.8,
			LearnFromEdits:  true,
		},
	}
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return defaults
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadDefault loads configuration from the default location
func LoadDefault() (*Config, error) {
	path := DefaultConfigPath()
	return Load(path)
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "lima", "config.yaml")
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	// Validate first
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate UI settings
	validViews := map[string]bool{
		"dashboard":    true,
		"transactions": true,
		"accounts":     true,
		"reports":      true,
	}
	if !validViews[c.UI.DefaultView] {
		return fmt.Errorf("invalid default view: %s", c.UI.DefaultView)
	}

	if c.UI.PageSize < 1 || c.UI.PageSize > 1000 {
		return fmt.Errorf("page size must be between 1 and 1000")
	}

	// Validate theme colors (basic check - should be hex colors)
	colors := []string{
		c.Theme.Primary,
		c.Theme.Secondary,
		c.Theme.Success,
		c.Theme.Warning,
		c.Theme.Error,
		c.Theme.Muted,
		c.Theme.Text,
		c.Theme.Background,
	}
	for _, color := range colors {
		if len(color) > 0 && color[0] != '#' {
			return fmt.Errorf("invalid color format (must start with #): %s", color)
		}
	}

	// Validate categorization settings
	if c.Categorization.ConfidenceThreshold < 0 || c.Categorization.ConfidenceThreshold > 1 {
		return fmt.Errorf("confidence threshold must be between 0 and 1")
	}

	// Validate keybindings (at least one key per action)
	keybindingFields := []struct {
		name string
		keys []string
	}{
		{"quit", c.Keybindings.Quit},
		{"dashboard", c.Keybindings.Dashboard},
		{"transactions", c.Keybindings.Transactions},
		{"accounts", c.Keybindings.Accounts},
	}

	for _, field := range keybindingFields {
		if len(field.keys) == 0 {
			return fmt.Errorf("keybinding '%s' must have at least one key", field.name)
		}
	}

	return nil
}

// Merge merges another config into this one (other takes precedence)
func (c *Config) Merge(other *Config) {
	// Merge files
	if other.Files.DefaultLedger != "" {
		c.Files.DefaultLedger = other.Files.DefaultLedger
	}
	if other.Files.PatternsFile != "" {
		c.Files.PatternsFile = other.Files.PatternsFile
	}

	// Merge UI
	if other.UI.DefaultView != "" {
		c.UI.DefaultView = other.UI.DefaultView
	}
	if other.UI.PageSize > 0 {
		c.UI.PageSize = other.UI.PageSize
	}
	if other.UI.DateFormat != "" {
		c.UI.DateFormat = other.UI.DateFormat
	}

	// Theme colors
	if other.Theme.Primary != "" {
		c.Theme.Primary = other.Theme.Primary
	}
	if other.Theme.Secondary != "" {
		c.Theme.Secondary = other.Theme.Secondary
	}

	// Keybindings - merge arrays
	if len(other.Keybindings.Quit) > 0 {
		c.Keybindings.Quit = other.Keybindings.Quit
	}
	if len(other.Keybindings.Dashboard) > 0 {
		c.Keybindings.Dashboard = other.Keybindings.Dashboard
	}
	if len(other.Keybindings.Transactions) > 0 {
		c.Keybindings.Transactions = other.Keybindings.Transactions
	}
	if len(other.Keybindings.Accounts) > 0 {
		c.Keybindings.Accounts = other.Keybindings.Accounts
	}
}
