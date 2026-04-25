// Package conf provides configuration management for BirdNET-Go.
// It handles loading, validation, and access to application settings.
package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// BirdNETConfig holds settings for the BirdNET analysis engine.
type BirdNETConfig struct {
	Sensitivity float64 `yaml:"sensitivity"` // Detection sensitivity (0.0-1.5)
	Threshold   float64 `yaml:"threshold"`   // Minimum confidence threshold (0.0-1.0)
	Overlap     float64 `yaml:"overlap"`     // Overlap between analysis chunks in seconds
	Locale      string  `yaml:"locale"`      // Locale for species labels (e.g., "en", "de")
	Latitude    float64 `yaml:"latitude"`    // Location latitude for species filtering
	Longitude   float64 `yaml:"longitude"`   // Location longitude for species filtering
}

// AudioConfig holds settings for audio capture and processing.
type AudioConfig struct {
	Source     string `yaml:"source"`     // Audio source (e.g., "sysdefault", "rtsp://...")
	Export     bool   `yaml:"export"`     // Whether to export audio clips of detections
	ExportPath string `yaml:"exportPath"` // Directory path for exported audio clips
	ExportType string `yaml:"exportType"` // Export format: "wav", "mp3", "flac"
}

// DatabaseConfig holds settings for the SQLite database.
type DatabaseConfig struct {
	Path string `yaml:"path"` // Path to the SQLite database file
	Type string `yaml:"type"` // Database type: "sqlite" (future: "postgres")
}

// ServerConfig holds settings for the web server.
type ServerConfig struct {
	Enabled bool   `yaml:"enabled"` // Whether the web server is enabled
	Port    int    `yaml:"port"`    // HTTP port to listen on
	Host    string `yaml:"host"`    // Host address to bind to
	LogFile string `yaml:"logFile"` // Path to server log file (empty = stdout)
}

// Config is the top-level application configuration structure.
type Config struct {
	BirdNET  BirdNETConfig  `yaml:"birdnet"`
	Audio    AudioConfig    `yaml:"audio"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	LogLevel string         `yaml:"logLevel"` // Logging level: "debug", "info", "warn", "error"
}

// DefaultConfig returns a Config populated with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		BirdNET: BirdNETConfig{
			Sensitivity: 1.0,
			Threshold:   0.75,
			Overlap:     1.5,
			Locale:      "en",
			Latitude:    0.0,
			Longitude:   0.0,
		},
		Audio: AudioConfig{
			Source:     "sysdefault",
			Export:     false,
			ExportPath: "clips",
			ExportType: "wav",
		},
		Database: DatabaseConfig{
			Path: "birdnet.db",
			Type: "sqlite",
		},
		Server: ServerConfig{
			Enabled: true,
			Port:    8080,
			Host:    "0.0.0.0",
		},
		LogLevel: "info",
	}
}

// Load reads a YAML configuration file from the given path and returns
// a Config. Missing fields are filled with defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults when no config file exists yet.
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that configuration values are within acceptable ranges.
func (c *Config) Validate() error {
	if c.BirdNET.Sensitivity < 0.0 || c.BirdNET.Sensitivity > 1.5 {
		return fmt.Errorf("birdnet.sensitivity must be between 0.0 and 1.5, got %.2f", c.BirdNET.Sensitivity)
	}
	if c.BirdNET.Threshold < 0.0 || c.BirdNET.Threshold > 1.0 {
		return fmt.Errorf("birdnet.threshold must be between 0.0 and 1.0, got %.2f", c.BirdNET.Threshold)
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
	}
	return nil
}
