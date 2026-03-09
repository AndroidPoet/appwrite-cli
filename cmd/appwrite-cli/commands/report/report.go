package report

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

// ReportCmd generates a full project report
var ReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a full project report",
	Long: `Generate a comprehensive report of your Appwrite project.

Includes: databases, collections, functions, storage, users, teams,
health status, and configuration summary.`,
	RunE: runReport,
}

type reportSection struct {
	Section string      `json:"section"`
	Data    interface{} `json:"data"`
}

func runReport(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}

	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		d = 120 * time.Second
	}

	client, err := api.NewClient(cli.GetProjectID(), d)
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	sections := []struct {
		name string
		path string
	}{
		{"Health", "/health"},
		{"Databases", "/databases?limit=100"},
		{"Functions", "/functions?limit=100"},
		{"Storage Buckets", "/storage/buckets?limit=100"},
		{"Users", "/users?limit=1"},
		{"Teams", "/teams?limit=100"},
	}

	report := make([]reportSection, 0, len(sections))

	for _, s := range sections {
		output.PrintInfo("Fetching %s...", s.name)
		var resp map[string]interface{}
		if err := client.Get(ctx, s.path, &resp); err != nil {
			report = append(report, reportSection{
				Section: s.name,
				Data:    map[string]string{"error": err.Error()},
			})
		} else {
			report = append(report, reportSection{
				Section: s.name,
				Data:    resp,
			})
		}
	}

	// Summary header
	fmt.Println()
	output.PrintSuccess("Report generated for project: %s", cli.GetProjectID())
	output.PrintInfo("Endpoint: %s", client.GetEndpoint())
	fmt.Println()

	return output.Print(report)
}
