package functions

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
	functionID  string
	name        string
	runtime     string
	entrypoint  string
	timeout     int
	enabled     bool
	logging     bool
	limit       int
	offset      int
	allPages    bool
)

// FunctionsCmd manages functions
var FunctionsCmd = &cobra.Command{
	Use:   "functions",
	Short: "Manage functions",
	Long:  `List, create, update, and delete serverless functions.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all functions",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get function details",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new function",
	RunE:  runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a function",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a function",
	RunE:  runDelete,
}

var listExecsCmd = &cobra.Command{
	Use:   "list-executions",
	Short: "List function executions",
	RunE:  runListExecutions,
}

var listVarsCmd = &cobra.Command{
	Use:   "list-variables",
	Short: "List function variables",
	RunE:  runListVariables,
}

var listDeploymentsCmd = &cobra.Command{
	Use:   "list-deployments",
	Short: "List function deployments",
	RunE:  runListDeployments,
}

func init() {
	listCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getCmd.Flags().StringVar(&functionID, "function-id", "", "function ID")
	getCmd.MarkFlagRequired("function-id")
	getCmd.RegisterFlagCompletionFunc("function-id", completion.FunctionIDs())

	createCmd.Flags().StringVar(&functionID, "function-id", "", "function ID (auto-generated if empty)")
	createCmd.Flags().StringVar(&name, "name", "", "function name")
	createCmd.Flags().StringVar(&runtime, "runtime", "", "runtime (e.g., node-18.0, python-3.9, go-1.21)")
	createCmd.Flags().StringVar(&entrypoint, "entrypoint", "", "entrypoint file")
	createCmd.Flags().IntVar(&timeout, "timeout", 15, "execution timeout in seconds")
	createCmd.Flags().BoolVar(&enabled, "enabled", true, "enable the function")
	createCmd.Flags().BoolVar(&logging, "logging", true, "enable logging")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("runtime")

	updateCmd.Flags().StringVar(&functionID, "function-id", "", "function ID")
	updateCmd.Flags().StringVar(&name, "name", "", "function name")
	updateCmd.Flags().StringVar(&runtime, "runtime", "", "runtime")
	updateCmd.Flags().StringVar(&entrypoint, "entrypoint", "", "entrypoint file")
	updateCmd.Flags().IntVar(&timeout, "timeout", 15, "execution timeout in seconds")
	updateCmd.Flags().BoolVar(&enabled, "enabled", true, "enable the function")
	updateCmd.Flags().BoolVar(&logging, "logging", true, "enable logging")
	updateCmd.MarkFlagRequired("function-id")
	updateCmd.MarkFlagRequired("name")
	updateCmd.RegisterFlagCompletionFunc("function-id", completion.FunctionIDs())

	var confirm bool
	deleteCmd.Flags().StringVar(&functionID, "function-id", "", "function ID")
	deleteCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("function-id")
	deleteCmd.RegisterFlagCompletionFunc("function-id", completion.FunctionIDs())

	for _, cmd := range []*cobra.Command{listExecsCmd, listVarsCmd, listDeploymentsCmd} {
		cmd.Flags().StringVar(&functionID, "function-id", "", "function ID")
		cmd.MarkFlagRequired("function-id")
		cmd.RegisterFlagCompletionFunc("function-id", completion.FunctionIDs())
	}
	listExecsCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listExecsCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")

	FunctionsCmd.AddCommand(listCmd)
	FunctionsCmd.AddCommand(getCmd)
	FunctionsCmd.AddCommand(createCmd)
	FunctionsCmd.AddCommand(updateCmd)
	FunctionsCmd.AddCommand(deleteCmd)
	FunctionsCmd.AddCommand(listExecsCmd)
	FunctionsCmd.AddCommand(listVarsCmd)
	FunctionsCmd.AddCommand(listDeploymentsCmd)
}

type FunctionInfo struct {
	ID         string `json:"$id"`
	Name       string `json:"name"`
	Runtime    string `json:"runtime"`
	Enabled    bool   `json:"enabled"`
	Entrypoint string `json:"entrypoint"`
	Timeout    int    `json:"timeout"`
	CreatedAt  string `json:"$createdAt,omitempty"`
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
		var items []FunctionInfo
		err := client.ListAll(ctx, "/functions", limit, "functions", func(raw json.RawMessage) error {
			var page []FunctionInfo
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
		Functions []FunctionInfo `json:"functions"`
	}
	path := fmt.Sprintf("/functions?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Functions)
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

	var fn FunctionInfo
	if err := client.Get(ctx, "/functions/"+functionID, &fn); err != nil {
		return err
	}
	return output.Print(fn)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create function '%s' with runtime '%s'", name, runtime)
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
		"runtime": runtime,
		"enabled": enabled,
		"logging": logging,
		"timeout": timeout,
	}
	if functionID != "" {
		body["functionId"] = functionID
	} else {
		body["functionId"] = "unique()"
	}
	if entrypoint != "" {
		body["entrypoint"] = entrypoint
	}

	var fn FunctionInfo
	if err := client.Post(ctx, "/functions", body, &fn); err != nil {
		return err
	}

	output.PrintSuccess("Function '%s' created (ID: %s)", name, fn.ID)
	return output.Print(fn)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update function '%s'", functionID)
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
		"logging": logging,
		"timeout": timeout,
	}
	if runtime != "" {
		body["runtime"] = runtime
	}
	if entrypoint != "" {
		body["entrypoint"] = entrypoint
	}

	var fn FunctionInfo
	if err := client.Put(ctx, "/functions/"+functionID, body, &fn); err != nil {
		return err
	}

	output.PrintSuccess("Function '%s' updated", functionID)
	return output.Print(fn)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete function '%s'", functionID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, "/functions/"+functionID); err != nil {
		return err
	}

	output.PrintSuccess("Function '%s' deleted", functionID)
	return nil
}

func runListExecutions(cmd *cobra.Command, args []string) error {
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
	path := fmt.Sprintf("/functions/%s/executions?limit=%d&offset=%d", functionID, limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runListVariables(cmd *cobra.Command, args []string) error {
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
	if err := client.Get(ctx, "/functions/"+functionID+"/variables", &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runListDeployments(cmd *cobra.Command, args []string) error {
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
	if err := client.Get(ctx, "/functions/"+functionID+"/deployments", &resp); err != nil {
		return err
	}
	return output.Print(resp)
}
