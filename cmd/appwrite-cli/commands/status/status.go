package status

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

// StatusCmd shows project overview
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project overview dashboard",
	Long: `Display a summary of your Appwrite project including databases,
collections, functions, storage buckets, users, and teams.`,
	RunE: runStatus,
}

type statusEntry struct {
	Resource string `json:"resource"`
	Count    int    `json:"count"`
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	entries := make([]statusEntry, 0, len(resources))
	for _, r := range resources {
		var resp map[string]interface{}
		if err := client.Get(ctx, r.path, &resp); err != nil {
			entries = append(entries, statusEntry{Resource: r.name, Count: -1})
			continue
		}
		total, _ := resp[r.countKey].(float64)
		entries = append(entries, statusEntry{Resource: r.name, Count: int(total)})
	}

	output.PrintInfo("Project: %s", cli.GetProjectID())
	output.PrintInfo("Endpoint: %s", client.GetEndpoint())
	fmt.Println()

	return output.Print(entries)
}
