package exportcmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

var (
	exportFile string
	importFile string
)

// ExportCmd exports project configuration to YAML
var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export project configuration to YAML",
	Long: `Export databases, collections, and their schemas to a YAML file.
This enables infrastructure-as-code workflows for Appwrite.

Examples:
  aw export --file appwrite.yaml
  aw export --file appwrite.yaml --project proj_123`,
	RunE: runExport,
}

// ImportCmd imports project configuration from YAML
var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import project configuration from YAML",
	Long: `Import databases and collections from a YAML configuration file.

Examples:
  aw import --file appwrite.yaml`,
	RunE: runImport,
}

func init() {
	ExportCmd.Flags().StringVar(&exportFile, "file", "appwrite.yaml", "output file path")
	ImportCmd.Flags().StringVar(&importFile, "file", "appwrite.yaml", "input file path")
}

type exportConfig struct {
	Version   string               `yaml:"version"`
	Project   string               `yaml:"project"`
	Databases []exportDatabase     `yaml:"databases"`
}

type exportDatabase struct {
	ID          string             `yaml:"id"`
	Name        string             `yaml:"name"`
	Collections []exportCollection `yaml:"collections,omitempty"`
}

type exportCollection struct {
	ID               string                   `yaml:"id"`
	Name             string                   `yaml:"name"`
	Enabled          bool                     `yaml:"enabled"`
	DocumentSecurity bool                     `yaml:"document_security"`
	Attributes       []map[string]interface{} `yaml:"attributes,omitempty"`
	Indexes          []map[string]interface{} `yaml:"indexes,omitempty"`
}

func runExport(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}

	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		d = 60 * time.Second
	}

	client, err := api.NewClient(cli.GetProjectID(), d)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	// Fetch databases
	var dbResp struct {
		Databases []struct {
			ID   string `json:"$id"`
			Name string `json:"name"`
		} `json:"databases"`
	}
	if err := client.Get(ctx, "/databases?limit=100", &dbResp); err != nil {
		return err
	}

	cfg := exportConfig{
		Version: "1.0",
		Project: cli.GetProjectID(),
	}

	for _, db := range dbResp.Databases {
		expDB := exportDatabase{
			ID:   db.ID,
			Name: db.Name,
		}

		// Fetch collections for this database
		var collResp struct {
			Collections []struct {
				ID               string `json:"$id"`
				Name             string `json:"name"`
				Enabled          bool   `json:"enabled"`
				DocumentSecurity bool   `json:"documentSecurity"`
			} `json:"collections"`
		}
		collPath := fmt.Sprintf("/databases/%s/collections?limit=100", db.ID)
		if err := client.Get(ctx, collPath, &collResp); err != nil {
			output.PrintWarning("Failed to fetch collections for database '%s': %v", db.ID, err)
			continue
		}

		for _, coll := range collResp.Collections {
			expColl := exportCollection{
				ID:               coll.ID,
				Name:             coll.Name,
				Enabled:          coll.Enabled,
				DocumentSecurity: coll.DocumentSecurity,
			}

			// Fetch attributes
			var attrResp struct {
				Attributes []map[string]interface{} `json:"attributes"`
			}
			attrPath := fmt.Sprintf("/databases/%s/collections/%s/attributes", db.ID, coll.ID)
			if err := client.Get(ctx, attrPath, &attrResp); err == nil {
				expColl.Attributes = attrResp.Attributes
			}

			// Fetch indexes
			var idxResp struct {
				Indexes []map[string]interface{} `json:"indexes"`
			}
			idxPath := fmt.Sprintf("/databases/%s/collections/%s/indexes", db.ID, coll.ID)
			if err := client.Get(ctx, idxPath, &idxResp); err == nil {
				expColl.Indexes = idxResp.Indexes
			}

			expDB.Collections = append(expDB.Collections, expColl)
		}

		cfg.Databases = append(cfg.Databases, expDB)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(exportFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	output.PrintSuccess("Exported to %s (%d databases)", exportFile, len(cfg.Databases))
	return nil
}

func runImport(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}

	data, err := os.ReadFile(importFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var cfg exportConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	d, parseErr := time.ParseDuration(cli.GetTimeout())
	if parseErr != nil {
		d = 60 * time.Second
	}

	client, err := api.NewClient(cli.GetProjectID(), d)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would import %d databases from %s", len(cfg.Databases), importFile)
		for _, db := range cfg.Databases {
			output.PrintInfo("  Database: %s (%d collections)", db.Name, len(db.Collections))
		}
		return nil
	}

	for _, db := range cfg.Databases {
		body := map[string]interface{}{
			"databaseId": db.ID,
			"name":       db.Name,
		}
		var dbResp map[string]interface{}
		if err := client.Post(ctx, "/databases", body, &dbResp); err != nil {
			output.PrintWarning("Database '%s': %v (may already exist)", db.Name, err)
		} else {
			output.PrintSuccess("Created database '%s'", db.Name)
		}

		for _, coll := range db.Collections {
			collBody := map[string]interface{}{
				"collectionId":     coll.ID,
				"name":             coll.Name,
				"enabled":          coll.Enabled,
				"documentSecurity": coll.DocumentSecurity,
			}
			collPath := fmt.Sprintf("/databases/%s/collections", db.ID)
			var collResp map[string]interface{}
			if err := client.Post(ctx, collPath, collBody, &collResp); err != nil {
				output.PrintWarning("Collection '%s': %v (may already exist)", coll.Name, err)
			} else {
				output.PrintSuccess("Created collection '%s' in '%s'", coll.Name, db.Name)
			}
		}
	}

	output.PrintSuccess("Import complete from %s", importFile)
	return nil
}
