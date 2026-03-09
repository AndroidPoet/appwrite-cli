package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Profile represents an authentication profile
type Profile struct {
	Name           string `json:"name"`
	APIKey         string `json:"api_key"`
	Endpoint       string `json:"endpoint,omitempty"`
	DefaultProject string `json:"default_project,omitempty"`
}

// Config represents the appwrite-cli configuration
type Config struct {
	DefaultProfile string             `json:"default_profile"`
	Profiles       map[string]Profile `json:"profiles"`
}

var (
	cfg            *Config
	currentProfile *Profile
	debugMode      bool
	configPath     string
)

// Init initializes the configuration
func Init(cfgFile, profileName string) error {
	if cfgFile != "" {
		configPath = cfgFile
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, ".appwrite-cli", "config.json")
	}

	cfg = &Config{
		DefaultProfile: "default",
		Profiles:       make(map[string]Profile),
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	if profileName == "" {
		profileName = viper.GetString("profile")
	}
	if profileName == "" {
		profileName = cfg.DefaultProfile
	}
	if profileName == "" {
		profileName = "default"
	}

	if p, ok := cfg.Profiles[profileName]; ok {
		currentProfile = &p
	} else {
		currentProfile = &Profile{Name: profileName}
	}

	// Override from environment
	if apiKey := viper.GetString("api_key"); apiKey != "" {
		currentProfile.APIKey = apiKey
	}
	if endpoint := viper.GetString("endpoint"); endpoint != "" {
		currentProfile.Endpoint = endpoint
	}

	return nil
}

// Save saves the current configuration
func Save() error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	return cfg
}

// GetProfile returns the current profile
func GetProfile() *Profile {
	return currentProfile
}

// SetProfile sets a profile in the configuration
func SetProfile(p Profile) {
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	cfg.Profiles[p.Name] = p
}

// DeleteProfile removes a profile from the configuration
func DeleteProfile(name string) {
	if cfg.Profiles != nil {
		delete(cfg.Profiles, name)
	}
}

// SetDefaultProfile sets the default profile name
func SetDefaultProfile(name string) {
	cfg.DefaultProfile = name
}

// GetAPIKey returns the API key for the current profile
func GetAPIKey() (string, error) {
	if currentProfile == nil {
		return "", fmt.Errorf("no profile configured. Run 'aw auth login' first")
	}

	if currentProfile.APIKey != "" {
		return currentProfile.APIKey, nil
	}

	return "", fmt.Errorf("no API key configured for profile '%s'. Run 'aw auth login' first", currentProfile.Name)
}

// GetEndpoint returns the endpoint for the current profile
func GetEndpoint() string {
	if currentProfile != nil && currentProfile.Endpoint != "" {
		return currentProfile.Endpoint
	}
	if ep := viper.GetString("endpoint"); ep != "" {
		return ep
	}
	return "https://cloud.appwrite.io/v1"
}

// SetDebug sets debug mode
func SetDebug(d bool) {
	debugMode = d
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	return debugMode
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	return configPath
}

// ListProfiles returns all profile names
func ListProfiles() []string {
	if cfg == nil || cfg.Profiles == nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	return names
}
