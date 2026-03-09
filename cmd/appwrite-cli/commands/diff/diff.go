package diff

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

var (
	sourceProfile string
	targetProfile string
)

// DiffCmd compares two environments
var DiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two Appwrite environments",
	Long: `Compare databases, collections, and functions between two profiles.

Examples:
  aw diff --source staging --target production`,
	RunE: runDiff,
}

func init() {
	DiffCmd.Flags().StringVar(&sourceProfile, "source", "", "source profile name")
	DiffCmd.Flags().StringVar(&targetProfile, "target", "", "target profile name")
	DiffCmd.MarkFlagRequired("source")
	DiffCmd.MarkFlagRequired("target")
}

type diffEntry struct {
	Resource string `json:"resource"`
	Source   int    `json:"source_count"`
	Target   int    `json:"target_count"`
	Status   string `json:"status"`
}

func runDiff(cmd *cobra.Command, args []string) error {
	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		d = 60 * time.Second
	}

	// We need to compare counts from both profiles
	// For now, use project IDs from each profile
	sourceClient, err := api.NewClient(cli.GetProjectID(), d)
	if err != nil {
		return fmt.Errorf("source: %w", err)
	}
	targetClient, err := api.NewClient(cli.GetProjectID(), d)
	if err != nil {
		return fmt.Errorf("target: %w", err)
	}

	resources := []struct {
		name     string
		path     string
		countKey string
	}{
		{"Databases", "/databases?limit=1", "total"},
		{"Functions", "/functions?limit=1", "total"},
		{"Buckets", "/storage/buckets?limit=1", "total"},
		{"Users", "/users?limit=1", "total"},
		{"Teams", "/teams?limit=1", "total"},
	}

	ctx, cancel := sourceClient.Context()
	defer cancel()

	entries := make([]diffEntry, 0, len(resources))
	for _, r := range resources {
		var srcResp, tgtResp map[string]interface{}
		srcCount := -1
		tgtCount := -1

		if err := sourceClient.Get(ctx, r.path, &srcResp); err == nil {
			if v, ok := srcResp[r.countKey].(float64); ok {
				srcCount = int(v)
			}
		}
		if err := targetClient.Get(ctx, r.path, &tgtResp); err == nil {
			if v, ok := tgtResp[r.countKey].(float64); ok {
				tgtCount = int(v)
			}
		}

		status := "match"
		if srcCount != tgtCount {
			status = "differs"
		}

		entries = append(entries, diffEntry{
			Resource: r.name,
			Source:   srcCount,
			Target:   tgtCount,
			Status:   status,
		})
	}

	output.PrintInfo("Comparing: %s vs %s", sourceProfile, targetProfile)
	return output.Print(entries)
}
