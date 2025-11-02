package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test defaults are set
	if cfg.UI.DefaultView != "dashboard" {
		t.Errorf("expected default view 'dashboard', got '%s'", cfg.UI.DefaultView)
	}

	if cfg.UI.PageSize != 20 {
		t.Errorf("expected page size 20, got %d", cfg.UI.PageSize)
	}

	if cfg.Theme.Primary != "#00D9FF" {
		t.Errorf("expected primary color '#00D9FF', got '%s'", cfg.Theme.Primary)
	}

	if len(cfg.Keybindings.Quit) == 0 {
		t.Error("expected quit keybinding to be set")
	}

	if cfg.Categorization.ConfidenceThreshold != 0.8 {
		t.Errorf("expected confidence threshold 0.8, got %f", cfg.Categorization.ConfidenceThreshold)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Config)
		shouldErr bool
	}{
		{
			name:      "valid config",
			mutate:    func(c *Config) {},
			shouldErr: false,
		},
		{
			name: "invalid default view",
			mutate: func(c *Config) {
				c.UI.DefaultView = "invalid"
			},
			shouldErr: true,
		},
		{
			name: "page size too small",
			mutate: func(c *Config) {
				c.UI.PageSize = 0
			},
			shouldErr: true,
		},
		{
			name: "page size too large",
			mutate: func(c *Config) {
				c.UI.PageSize = 1001
			},
			shouldErr: true,
		},
		{
			name: "invalid color format",
			mutate: func(c *Config) {
				c.Theme.Primary = "blue"
			},
			shouldErr: true,
		},
		{
			name: "confidence threshold too low",
			mutate: func(c *Config) {
				c.Categorization.ConfidenceThreshold = -0.1
			},
			shouldErr: true,
		},
		{
			name: "confidence threshold too high",
			mutate: func(c *Config) {
				c.Categorization.ConfidenceThreshold = 1.1
			},
			shouldErr: true,
		},
		{
			name: "missing quit keybinding",
			mutate: func(c *Config) {
				c.Keybindings.Quit = []string{}
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(cfg)

			err := cfg.Validate()
			if tt.shouldErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestConfigLoadSave(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a config
	cfg := DefaultConfig()
	cfg.UI.DefaultView = "transactions"
	cfg.UI.PageSize = 50
	cfg.Theme.Primary = "#FF0000"

	// Save it
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify values
	if loaded.UI.DefaultView != "transactions" {
		t.Errorf("expected default view 'transactions', got '%s'", loaded.UI.DefaultView)
	}

	if loaded.UI.PageSize != 50 {
		t.Errorf("expected page size 50, got %d", loaded.UI.PageSize)
	}

	if loaded.Theme.Primary != "#FF0000" {
		t.Errorf("expected primary color '#FF0000', got '%s'", loaded.Theme.Primary)
	}
}

func TestConfigLoadNonExistent(t *testing.T) {
	// Try to load non-existent file
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got: %v", err)
	}

	// Should return defaults
	if cfg.UI.DefaultView != "dashboard" {
		t.Error("expected default config when file doesn't exist")
	}
}

func TestConfigLoadInvalid(t *testing.T) {
	// Create temporary file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `this is not: valid: yaml: content`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Try to load
	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestConfigMerge(t *testing.T) {
	base := DefaultConfig()
	base.UI.DefaultView = "dashboard"
	base.UI.PageSize = 20
	base.Theme.Primary = "#00D9FF"

	override := &Config{
		UI: UIConfig{
			DefaultView: "transactions",
			PageSize:    50,
		},
		Theme: ThemeConfig{
			Primary: "#FF0000",
		},
	}

	base.Merge(override)

	// Check merged values
	if base.UI.DefaultView != "transactions" {
		t.Errorf("expected merged default view 'transactions', got '%s'", base.UI.DefaultView)
	}

	if base.UI.PageSize != 50 {
		t.Errorf("expected merged page size 50, got %d", base.UI.PageSize)
	}

	if base.Theme.Primary != "#FF0000" {
		t.Errorf("expected merged primary color '#FF0000', got '%s'", base.Theme.Primary)
	}

	// Check non-overridden values remain
	if base.Theme.Secondary != "#7D56F4" {
		t.Error("non-overridden value should remain unchanged")
	}
}

func TestConfigPartialLoad(t *testing.T) {
	// Create temporary file with partial config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	partialYAML := `ui:
  default_view: transactions
  page_size: 30
theme:
  primary: "#FF0000"
`
	if err := os.WriteFile(configPath, []byte(partialYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Check overridden values
	if cfg.UI.DefaultView != "transactions" {
		t.Errorf("expected default view 'transactions', got '%s'", cfg.UI.DefaultView)
	}

	if cfg.UI.PageSize != 30 {
		t.Errorf("expected page size 30, got %d", cfg.UI.PageSize)
	}

	if cfg.Theme.Primary != "#FF0000" {
		t.Errorf("expected primary color '#FF0000', got '%s'", cfg.Theme.Primary)
	}

	// Check defaults are preserved for non-specified values
	if cfg.Theme.Secondary != "#7D56F4" {
		t.Error("non-specified value should use default")
	}

	if len(cfg.Keybindings.Quit) == 0 {
		t.Error("non-specified keybindings should use defaults")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()

	if path == "" {
		t.Error("default config path should not be empty")
	}

	// Should end with .config/lima/config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected path to end with 'config.yaml', got '%s'", path)
	}
}
