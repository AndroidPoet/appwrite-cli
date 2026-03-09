package initcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

const configFileName = ".aw.yaml"

// ProjectConfig represents a project-level configuration file
type ProjectConfig struct {
	Project  string `yaml:"project" json:"project"`
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Output   string `yaml:"output,omitempty" json:"output,omitempty"`
}

var (
	initProject  string
	initEndpoint string
	initOutput   string
	initForce    bool
)

// InitCmd creates a project configuration file
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project configuration",
	Long: `Create a .aw.yaml configuration file in the current directory.

This file stores project-level defaults so you don't have to specify
them on every command.`,
	RunE: runInit,
}

func init() {
	InitCmd.Flags().StringVar(&initProject, "project", "", "Appwrite project ID")
	InitCmd.Flags().StringVar(&initEndpoint, "endpoint", "", "Appwrite endpoint URL")
	InitCmd.Flags().StringVar(&initOutput, "output", "json", "default output format")
	InitCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing config")
	InitCmd.MarkFlagRequired("project")
}

func runInit(cmd *cobra.Command, args []string) error {
	configFile := filepath.Join(".", configFileName)

	if _, err := os.Stat(configFile); err == nil && !initForce {
		return fmt.Errorf("%s already exists. Use --force to overwrite", configFileName)
	}

	cfg := ProjectConfig{
		Project:  initProject,
		Endpoint: initEndpoint,
		Output:   initOutput,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configFileName, err)
	}

	output.PrintSuccess("Created %s", configFile)
	return output.Print(cfg)
}

// FindProjectConfig searches for .aw.yaml in the current and parent directories
func FindProjectConfig(startDir string) string {
	dir := startDir
	for {
		configFile := filepath.Join(dir, configFileName)
		if _, err := os.Stat(configFile); err == nil {
			return configFile
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
