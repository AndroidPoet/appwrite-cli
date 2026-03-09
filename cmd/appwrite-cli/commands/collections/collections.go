package collections

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
	databaseID      string
	collectionID    string
	name            string
	enabled         bool
	documentSec     bool
	limit           int
	offset          int
	allPages        bool
)

// CollectionsCmd manages collections
var CollectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Manage collections",
	Long:  `List, create, update, and delete collections within a database.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all collections in a database",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get collection details",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new collection",
	RunE:  runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a collection",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a collection",
	RunE:  runDelete,
}

var listAttrsCmd = &cobra.Command{
	Use:   "list-attributes",
	Short: "List collection attributes",
	RunE:  runListAttributes,
}

var listIndexesCmd = &cobra.Command{
	Use:   "list-indexes",
	Short: "List collection indexes",
	RunE:  runListIndexes,
}

func init() {
	// Shared database-id flag
	for _, cmd := range []*cobra.Command{listCmd, getCmd, createCmd, updateCmd, deleteCmd, listAttrsCmd, listIndexesCmd} {
		cmd.Flags().StringVar(&databaseID, "database-id", "", "database ID")
		cmd.MarkFlagRequired("database-id")
		cmd.RegisterFlagCompletionFunc("database-id", completion.DatabaseIDs())
	}

	listCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getCmd.Flags().StringVar(&collectionID, "collection-id", "", "collection ID")
	getCmd.MarkFlagRequired("collection-id")

	createCmd.Flags().StringVar(&collectionID, "collection-id", "", "collection ID (auto-generated if empty)")
	createCmd.Flags().StringVar(&name, "name", "", "collection name")
	createCmd.Flags().BoolVar(&enabled, "enabled", true, "enable the collection")
	createCmd.Flags().BoolVar(&documentSec, "document-security", false, "enable document-level security")
	createCmd.MarkFlagRequired("name")

	updateCmd.Flags().StringVar(&collectionID, "collection-id", "", "collection ID")
	updateCmd.Flags().StringVar(&name, "name", "", "new collection name")
	updateCmd.Flags().BoolVar(&enabled, "enabled", true, "enable/disable the collection")
	updateCmd.Flags().BoolVar(&documentSec, "document-security", false, "enable/disable document-level security")
	updateCmd.MarkFlagRequired("collection-id")
	updateCmd.MarkFlagRequired("name")

	var confirm bool
	deleteCmd.Flags().StringVar(&collectionID, "collection-id", "", "collection ID")
	deleteCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("collection-id")

	listAttrsCmd.Flags().StringVar(&collectionID, "collection-id", "", "collection ID")
	listAttrsCmd.MarkFlagRequired("collection-id")

	listIndexesCmd.Flags().StringVar(&collectionID, "collection-id", "", "collection ID")
	listIndexesCmd.MarkFlagRequired("collection-id")

	CollectionsCmd.AddCommand(listCmd)
	CollectionsCmd.AddCommand(getCmd)
	CollectionsCmd.AddCommand(createCmd)
	CollectionsCmd.AddCommand(updateCmd)
	CollectionsCmd.AddCommand(deleteCmd)
	CollectionsCmd.AddCommand(listAttrsCmd)
	CollectionsCmd.AddCommand(listIndexesCmd)
}

type CollectionInfo struct {
	ID               string `json:"$id"`
	DatabaseID       string `json:"databaseId"`
	Name             string `json:"name"`
	Enabled          bool   `json:"enabled"`
	DocumentSecurity bool   `json:"documentSecurity"`
	CreatedAt        string `json:"$createdAt,omitempty"`
	UpdatedAt        string `json:"$updatedAt,omitempty"`
}

func parseTimeout() time.Duration {
	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		return 60 * time.Second
	}
	return d
}

func basePath() string {
	return fmt.Sprintf("/databases/%s/collections", databaseID)
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
		var items []CollectionInfo
		err := client.ListAll(ctx, basePath(), limit, "collections", func(raw json.RawMessage) error {
			var page []CollectionInfo
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
		Total       int              `json:"total"`
		Collections []CollectionInfo `json:"collections"`
	}
	path := fmt.Sprintf("%s?limit=%d&offset=%d", basePath(), limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Collections)
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

	var coll CollectionInfo
	if err := client.Get(ctx, basePath()+"/"+collectionID, &coll); err != nil {
		return err
	}
	return output.Print(coll)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create collection '%s' in database '%s'", name, databaseID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name":             name,
		"enabled":          enabled,
		"documentSecurity": documentSec,
	}
	if collectionID != "" {
		body["collectionId"] = collectionID
	} else {
		body["collectionId"] = "unique()"
	}

	var coll CollectionInfo
	if err := client.Post(ctx, basePath(), body, &coll); err != nil {
		return err
	}

	output.PrintSuccess("Collection '%s' created (ID: %s)", name, coll.ID)
	return output.Print(coll)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update collection '%s'", collectionID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name":             name,
		"enabled":          enabled,
		"documentSecurity": documentSec,
	}

	var coll CollectionInfo
	if err := client.Put(ctx, basePath()+"/"+collectionID, body, &coll); err != nil {
		return err
	}

	output.PrintSuccess("Collection '%s' updated", collectionID)
	return output.Print(coll)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete collection '%s'", collectionID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, basePath()+"/"+collectionID); err != nil {
		return err
	}

	output.PrintSuccess("Collection '%s' deleted", collectionID)
	return nil
}

func runListAttributes(cmd *cobra.Command, args []string) error {
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
	path := fmt.Sprintf("%s/%s/attributes", basePath(), collectionID)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runListIndexes(cmd *cobra.Command, args []string) error {
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
	path := fmt.Sprintf("%s/%s/indexes", basePath(), collectionID)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}
