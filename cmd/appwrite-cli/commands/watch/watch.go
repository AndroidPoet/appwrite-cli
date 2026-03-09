package watch

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

var interval int

// WatchCmd provides live monitoring
var WatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Live monitoring dashboard",
	Long:  `Watch your Appwrite project metrics in real-time.`,
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Watch health status",
	RunE:  runWatchHealth,
}

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Watch project metrics",
	RunE:  runWatchMetrics,
}

func init() {
	WatchCmd.PersistentFlags().IntVar(&interval, "interval", 5, "refresh interval in seconds")

	WatchCmd.AddCommand(healthCmd)
	WatchCmd.AddCommand(metricsCmd)
}

func runWatchHealth(cmd *cobra.Command, args []string) error {
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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	output.PrintInfo("Watching health (every %ds, Ctrl+C to stop)...\n", interval)

	for {
		ctx, cancel := client.Context()
		var resp map[string]interface{}
		if err := client.Get(ctx, "/health", &resp); err != nil {
			output.PrintError("Health check failed: %v", err)
		} else {
			fmt.Printf("\033[2J\033[H") // Clear screen
			output.PrintInfo("Health Status — %s", time.Now().Format("15:04:05"))
			output.PrintInfo("Project: %s", cli.GetProjectID())
			fmt.Println()
			output.Print(resp)
		}
		cancel()

		select {
		case <-ticker.C:
			continue
		case <-sig:
			fmt.Println()
			output.PrintInfo("Stopped watching.")
			return nil
		}
	}
}

func runWatchMetrics(cmd *cobra.Command, args []string) error {
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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	output.PrintInfo("Watching metrics (every %ds, Ctrl+C to stop)...\n", interval)

	resources := []struct {
		name string
		path string
	}{
		{"Databases", "/databases?limit=1"},
		{"Functions", "/functions?limit=1"},
		{"Buckets", "/storage/buckets?limit=1"},
		{"Users", "/users?limit=1"},
		{"Teams", "/teams?limit=1"},
	}

	for {
		ctx, cancel := client.Context()
		fmt.Printf("\033[2J\033[H")
		output.PrintInfo("Project Metrics — %s", time.Now().Format("15:04:05"))
		output.PrintInfo("Project: %s", cli.GetProjectID())
		fmt.Println()

		for _, r := range resources {
			var resp map[string]interface{}
			if err := client.Get(ctx, r.path, &resp); err != nil {
				fmt.Printf("  %-15s  error\n", r.name)
			} else {
				total, _ := resp["total"].(float64)
				fmt.Printf("  %-15s  %d\n", r.name, int(total))
			}
		}
		cancel()

		select {
		case <-ticker.C:
			continue
		case <-sig:
			fmt.Println()
			output.PrintInfo("Stopped watching.")
			return nil
		}
	}
}
