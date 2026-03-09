package documents

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
	databaseID   string
	collectionID string
	documentID   string
	data         string
	permissions  []string
	limit        int
	offset       int
	allPages     bool
)

// DocumentsCmd manages documents
var DocumentsCmd = &cobra.Command{
	Use:   "documents",
	Short: "Manage documents",
	Long:  `List, create, update, and delete documents within a collection.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documents in a collection",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a document",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new document",
	Long: `Create a new document in a collection.

The --data flag accepts a JSON string with document fields.

Examples:
  aw documents create --database-id db1 --collection-id col1 --data '{"title":"Hello","body":"World"}'`,
	RunE: runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a document",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a document",
	RunE:  runDelete,
}

func init() {
	for _, cmd := range []*cobra.Command{listCmd, getCmd, createCmd, updateCmd, deleteCmd} {
		cmd.Flags().StringVar(&databaseID, "database-id", "", "database ID")
		cmd.Flags().StringVar(&collectionID, "collection-id", "", "collection ID")
		cmd.MarkFlagRequired("database-id")
		cmd.MarkFlagRequired("collection-id")
		cmd.RegisterFlagCompletionFunc("database-id", completion.DatabaseIDs())
	}

	listCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getCmd.Flags().StringVar(&documentID, "document-id", "", "document ID")
	getCmd.MarkFlagRequired("document-id")

	createCmd.Flags().StringVar(&documentID, "document-id", "", "document ID (auto-generated if empty)")
	createCmd.Flags().StringVar(&data, "data", "", "document data as JSON")
	createCmd.Flags().StringSliceVar(&permissions, "permissions", nil, "document permissions")
	createCmd.MarkFlagRequired("data")

	updateCmd.Flags().StringVar(&documentID, "document-id", "", "document ID")
	updateCmd.Flags().StringVar(&data, "data", "", "document data as JSON")
	updateCmd.Flags().StringSliceVar(&permissions, "permissions", nil, "document permissions")
	updateCmd.MarkFlagRequired("document-id")
	updateCmd.MarkFlagRequired("data")

	var confirm bool
	deleteCmd.Flags().StringVar(&documentID, "document-id", "", "document ID")
	deleteCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("document-id")

	DocumentsCmd.AddCommand(listCmd)
	DocumentsCmd.AddCommand(getCmd)
	DocumentsCmd.AddCommand(createCmd)
	DocumentsCmd.AddCommand(updateCmd)
	DocumentsCmd.AddCommand(deleteCmd)
}

func parseTimeout() time.Duration {
	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		return 60 * time.Second
	}
	return d
}

func basePath() string {
	return fmt.Sprintf("/databases/%s/collections/%s/documents", databaseID, collectionID)
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
		var items []map[string]interface{}
		err := client.ListAll(ctx, basePath(), limit, "documents", func(raw json.RawMessage) error {
			var page []map[string]interface{}
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
		Total     int                      `json:"total"`
		Documents []map[string]interface{} `json:"documents"`
	}
	path := fmt.Sprintf("%s?limit=%d&offset=%d", basePath(), limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Documents)
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

	var doc map[string]interface{}
	if err := client.Get(ctx, basePath()+"/"+documentID, &doc); err != nil {
		return err
	}
	return output.Print(doc)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create document in collection '%s'", collectionID)
		return nil
	}

	var docData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &docData); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"data": docData,
	}
	if documentID != "" {
		body["documentId"] = documentID
	} else {
		body["documentId"] = "unique()"
	}
	if len(permissions) > 0 {
		body["permissions"] = permissions
	}

	var doc map[string]interface{}
	if err := client.Post(ctx, basePath(), body, &doc); err != nil {
		return err
	}

	output.PrintSuccess("Document created (ID: %s)", doc["$id"])
	return output.Print(doc)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update document '%s'", documentID)
		return nil
	}

	var docData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &docData); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"data": docData,
	}
	if len(permissions) > 0 {
		body["permissions"] = permissions
	}

	var doc map[string]interface{}
	if err := client.Patch(ctx, basePath()+"/"+documentID, body, &doc); err != nil {
		return err
	}

	output.PrintSuccess("Document '%s' updated", documentID)
	return output.Print(doc)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete document '%s'", documentID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, basePath()+"/"+documentID); err != nil {
		return err
	}

	output.PrintSuccess("Document '%s' deleted", documentID)
	return nil
}
