package databases

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/completion"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

var (
	databaseID string
	name       string
	enabled    bool
	limit      int
	offset     int
	allPages   bool
)

// DatabasesCmd manages databases
var DatabasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "Manage databases",
	Long:  `List, create, update, and delete Appwrite databases.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all databases",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get database details",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new database",
	RunE:  runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a database",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a database",
	RunE:  runDelete,
}

func init() {
	listCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getCmd.Flags().StringVar(&databaseID, "database-id", "", "database ID")
	getCmd.MarkFlagRequired("database-id")
	getCmd.RegisterFlagCompletionFunc("database-id", completion.DatabaseIDs())

	createCmd.Flags().StringVar(&databaseID, "database-id", "", "database ID (auto-generated if empty)")
	createCmd.Flags().StringVar(&name, "name", "", "database name")
	createCmd.Flags().BoolVar(&enabled, "enabled", true, "enable the database")
	createCmd.MarkFlagRequired("name")

	updateCmd.Flags().StringVar(&databaseID, "database-id", "", "database ID")
	updateCmd.Flags().StringVar(&name, "name", "", "new database name")
	updateCmd.Flags().BoolVar(&enabled, "enabled", true, "enable/disable the database")
	updateCmd.MarkFlagRequired("database-id")
	updateCmd.MarkFlagRequired("name")
	updateCmd.RegisterFlagCompletionFunc("database-id", completion.DatabaseIDs())

	var confirm bool
	deleteCmd.Flags().StringVar(&databaseID, "database-id", "", "database ID")
	deleteCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("database-id")
	deleteCmd.RegisterFlagCompletionFunc("database-id", completion.DatabaseIDs())

	DatabasesCmd.AddCommand(listCmd)
	DatabasesCmd.AddCommand(getCmd)
	DatabasesCmd.AddCommand(createCmd)
	DatabasesCmd.AddCommand(updateCmd)
	DatabasesCmd.AddCommand(deleteCmd)
}

type DatabaseInfo struct {
	ID        string `json:"$id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"$createdAt,omitempty"`
	UpdatedAt string `json:"$updatedAt,omitempty"`
}

func parseTimeout() time.Duration {
	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		return 60 * time.Second
	}
	return d
}

func runList(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}

	ctx, cancel := client.Context()
	defer cancel()

	if allPages {
		var items []DatabaseInfo
		err := client.ListAll(ctx, "/databases", limit, "databases", func(raw json.RawMessage) error {
			var page []DatabaseInfo
			if err := json.Unmarshal(raw, &page); err != nil {
				return err
			}
			items = append(items, page...)
			return nil
		})
		if err != nil {
			return err
		}
		return output.Print(items)
	}

	var resp struct {
		Total     int            `json:"total"`
		Databases []DatabaseInfo `json:"databases"`
	}
	path := fmt.Sprintf("/databases?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page, or --all for everything.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Databases)
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

	var db DatabaseInfo
	if err := client.Get(ctx, "/databases/"+databaseID, &db); err != nil {
		return err
	}
	return output.Print(db)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create database '%s'", name)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name":    name,
		"enabled": enabled,
	}
	if databaseID != "" {
		body["databaseId"] = databaseID
	} else {
		body["databaseId"] = "unique()"
	}

	var db DatabaseInfo
	if err := client.Post(ctx, "/databases", body, &db); err != nil {
		return err
	}

	output.PrintSuccess("Database '%s' created (ID: %s)", name, db.ID)
	return output.Print(db)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update database '%s'", databaseID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name":    name,
		"enabled": enabled,
	}

	var db DatabaseInfo
	if err := client.Put(ctx, "/databases/"+databaseID, body, &db); err != nil {
		return err
	}

	output.PrintSuccess("Database '%s' updated", databaseID)
	return output.Print(db)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete database '%s'", databaseID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, "/databases/"+databaseID); err != nil {
		return err
	}

	output.PrintSuccess("Database '%s' deleted", databaseID)
	return nil
}
