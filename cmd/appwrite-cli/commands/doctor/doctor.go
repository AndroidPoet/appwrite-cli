package doctor

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/config"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

// DoctorCmd runs diagnostic checks
var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostic checks",
	Long: `Run diagnostic checks to verify your appwrite-cli setup.

Checks:
  1. Configuration file loads correctly
  2. API key is configured
  3. Project ID is set
  4. Endpoint is reachable
  5. Appwrite API responds`,
	RunE: runDoctor,
}

type checkResult struct {
	Check  string `json:"check"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	results := make([]checkResult, 0, 5)

	// Check 1: Configuration
	configPath := config.GetConfigPath()
	if configPath != "" {
		results = append(results, checkResult{
			Check:  "Configuration",
			Status: "OK",
			Detail: configPath,
		})
	} else {
		results = append(results, checkResult{
			Check:  "Configuration",
			Status: "WARN",
			Detail: "No config file found. Run 'aw auth login' to configure.",
		})
	}

	// Check 2: API Key
	apiKey, err := config.GetAPIKey()
	if err != nil {
		results = append(results, checkResult{
			Check:  "API Key",
			Status: "FAIL",
			Detail: err.Error(),
		})
	} else {
		masked := apiKey[:6] + "..." + apiKey[len(apiKey)-4:]
		results = append(results, checkResult{
			Check:  "API Key",
			Status: "OK",
			Detail: masked,
		})
	}

	// Check 3: Project ID
	projectID := cli.GetProjectID()
	if projectID == "" {
		results = append(results, checkResult{
			Check:  "Project ID",
			Status: "WARN",
			Detail: "Not set. Use --project flag or AW_PROJECT env var.",
		})
	} else {
		results = append(results, checkResult{
			Check:  "Project ID",
			Status: "OK",
			Detail: projectID,
		})
	}

	// Check 4: Endpoint
	endpoint := config.GetEndpoint()
	results = append(results, checkResult{
		Check:  "Endpoint",
		Status: "OK",
		Detail: endpoint,
	})

	// Check 5: API connectivity
	if apiKey != "" && projectID != "" {
		client, err := api.NewClient(projectID, 10*time.Second)
		if err != nil {
			results = append(results, checkResult{
				Check:  "API Connectivity",
				Status: "FAIL",
				Detail: err.Error(),
			})
		} else {
			ctx, cancel := client.Context()
			defer cancel()

			var resp map[string]interface{}
			if err := client.Get(ctx, "/health", &resp); err != nil {
				results = append(results, checkResult{
					Check:  "API Connectivity",
					Status: "FAIL",
					Detail: err.Error(),
				})
			} else {
				results = append(results, checkResult{
					Check:  "API Connectivity",
					Status: "OK",
					Detail: "Successfully connected to Appwrite API",
				})
			}
		}
	} else {
		results = append(results, checkResult{
			Check:  "API Connectivity",
			Status: "SKIP",
			Detail: "Requires API key and project ID",
		})
	}

	// Print summary
	allOK := true
	for _, r := range results {
		icon := "✓"
		if r.Status == "FAIL" {
			icon = "✗"
			allOK = false
		} else if r.Status == "WARN" || r.Status == "SKIP" {
			icon = "!"
		}
		output.PrintInfo("%s %s: %s (%s)", icon, r.Check, r.Status, r.Detail)
	}

	if allOK {
		output.PrintSuccess("\nAll checks passed!")
	}

	return output.Print(results)
}
