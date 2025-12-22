package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDir  = ".afk"
	configFile = "config.json"
)

// OutputFormat defines the output style
type OutputFormat string

const (
	FormatLLM   OutputFormat = "llm"   // Structured for LLM agents (default)
	FormatHuman OutputFormat = "human" // Pretty output for humans
	FormatJSON  OutputFormat = "json"  // Pure JSON output
)

// Config holds the stored credentials and settings
type Config struct {
	APIKey           string       `json:"api_key"`
	APIURL           string       `json:"api_url"`
	SysName          string       `json:"sys_name,omitempty"`          // Name of the AI agent/system (e.g., "Claude Code")
	ReminderInterval string       `json:"reminder_interval,omitempty"` // e.g., "15m", "0" to disable
	Format           OutputFormat `json:"format,omitempty"`            // llm, human, json
}

// DefaultAPIURL is the default ChatBridge API endpoint
const DefaultAPIURL = "https://chatbridge.net"

// DevAPIURL is the development API endpoint
const DevAPIURL = "https://dev.chatbridge.net"

// configPath returns the full path to the config file
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load reads the config from ~/.afk/config.json
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in: run 'afk login' first")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("invalid config: missing API key")
	}

	// Default API URL if not set
	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}

	// Default sys_name
	if cfg.SysName == "" {
		cfg.SysName = "AI Agent"
	}

	// Default reminder interval (15 minutes)
	if cfg.ReminderInterval == "" {
		cfg.ReminderInterval = "15m"
	}

	// Default format (LLM)
	if cfg.Format == "" {
		cfg.Format = FormatLLM
	}

	return &cfg, nil
}

// Save writes the config to ~/.afk/config.json
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with restricted permissions (user read/write only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Delete removes the config file (logout)
func Delete() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // Already logged out
		}
		return fmt.Errorf("failed to remove config: %w", err)
	}

	return nil
}

// Exists checks if a config file exists
func Exists() bool {
	path, err := configPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Path returns the config file path for display
func Path() string {
	path, err := configPath()
	if err != nil {
		return "~/.afk/config.json"
	}
	return path
}
