package health

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

// HealthCmd checks Appwrite service health
var HealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check Appwrite service health",
	Long:  `Check the health status of your Appwrite instance.`,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get overall health status",
	RunE:  runGet,
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Check database health",
	RunE:  runDB,
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Check cache health",
	RunE:  runCache,
}

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Check storage health",
	RunE:  runStorage,
}

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Check queue health",
	RunE:  runQueue,
}

func init() {
	HealthCmd.AddCommand(getCmd)
	HealthCmd.AddCommand(dbCmd)
	HealthCmd.AddCommand(cacheCmd)
	HealthCmd.AddCommand(storageCmd)
	HealthCmd.AddCommand(queueCmd)
}

func parseTimeout() time.Duration {
	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		return 60 * time.Second
	}
	return d
}

func runGet(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	var resp map[string]interface{}
	if err := client.Get(ctx, "/health", &resp); err != nil {
		return err
	}

	output.PrintSuccess("Appwrite instance is healthy")
	return output.Print(resp)
}

func runDB(cmd *cobra.Command, args []string) error {
	return checkHealth("/health/db", "Database")
}

func runCache(cmd *cobra.Command, args []string) error {
	return checkHealth("/health/cache", "Cache")
}

func runStorage(cmd *cobra.Command, args []string) error {
	return checkHealth("/health/storage/local", "Storage")
}

func runQueue(cmd *cobra.Command, args []string) error {
	return checkHealth("/health/queue", "Queue")
}

func checkHealth(path, name string) error {
	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	var resp map[string]interface{}
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	output.PrintSuccess("%s is healthy", name)
	return output.Print(resp)
}
