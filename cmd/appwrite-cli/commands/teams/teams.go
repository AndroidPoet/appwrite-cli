package teams

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
	teamID   string
	name     string
	roles    []string
	limit    int
	offset   int
	allPages bool
)

// TeamsCmd manages teams
var TeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage teams",
	Long:  `List, create, update, and delete teams and memberships.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get team details",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new team",
	RunE:  runCreate,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a team",
	RunE:  runUpdate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a team",
	RunE:  runDelete,
}

var listMembersCmd = &cobra.Command{
	Use:   "list-members",
	Short: "List team memberships",
	RunE:  runListMembers,
}

func init() {
	listCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getCmd.Flags().StringVar(&teamID, "team-id", "", "team ID")
	getCmd.MarkFlagRequired("team-id")
	getCmd.RegisterFlagCompletionFunc("team-id", completion.TeamIDs())

	createCmd.Flags().StringVar(&teamID, "team-id", "", "team ID (auto-generated if empty)")
	createCmd.Flags().StringVar(&name, "name", "", "team name")
	createCmd.Flags().StringSliceVar(&roles, "roles", nil, "team roles")
	createCmd.MarkFlagRequired("name")

	updateCmd.Flags().StringVar(&teamID, "team-id", "", "team ID")
	updateCmd.Flags().StringVar(&name, "name", "", "new team name")
	updateCmd.MarkFlagRequired("team-id")
	updateCmd.MarkFlagRequired("name")
	updateCmd.RegisterFlagCompletionFunc("team-id", completion.TeamIDs())

	var confirm bool
	deleteCmd.Flags().StringVar(&teamID, "team-id", "", "team ID")
	deleteCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("team-id")
	deleteCmd.RegisterFlagCompletionFunc("team-id", completion.TeamIDs())

	listMembersCmd.Flags().StringVar(&teamID, "team-id", "", "team ID")
	listMembersCmd.MarkFlagRequired("team-id")
	listMembersCmd.RegisterFlagCompletionFunc("team-id", completion.TeamIDs())

	TeamsCmd.AddCommand(listCmd)
	TeamsCmd.AddCommand(getCmd)
	TeamsCmd.AddCommand(createCmd)
	TeamsCmd.AddCommand(updateCmd)
	TeamsCmd.AddCommand(deleteCmd)
	TeamsCmd.AddCommand(listMembersCmd)
}

type TeamInfo struct {
	ID        string `json:"$id"`
	Name      string `json:"name"`
	Total     int    `json:"total"`
	CreatedAt string `json:"$createdAt,omitempty"`
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
		var items []TeamInfo
		err := client.ListAll(ctx, "/teams", limit, "teams", func(raw json.RawMessage) error {
			var page []TeamInfo
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
		Total int        `json:"total"`
		Teams []TeamInfo `json:"teams"`
	}
	path := fmt.Sprintf("/teams?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Teams)
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

	var team TeamInfo
	if err := client.Get(ctx, "/teams/"+teamID, &team); err != nil {
		return err
	}
	return output.Print(team)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create team '%s'", name)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name": name,
	}
	if teamID != "" {
		body["teamId"] = teamID
	} else {
		body["teamId"] = "unique()"
	}
	if len(roles) > 0 {
		body["roles"] = roles
	}

	var team TeamInfo
	if err := client.Post(ctx, "/teams", body, &team); err != nil {
		return err
	}

	output.PrintSuccess("Team '%s' created (ID: %s)", name, team.ID)
	return output.Print(team)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update team '%s'", teamID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name": name,
	}

	var team TeamInfo
	if err := client.Put(ctx, "/teams/"+teamID, body, &team); err != nil {
		return err
	}

	output.PrintSuccess("Team '%s' updated", teamID)
	return output.Print(team)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete team '%s'", teamID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, "/teams/"+teamID); err != nil {
		return err
	}

	output.PrintSuccess("Team '%s' deleted", teamID)
	return nil
}

func runListMembers(cmd *cobra.Command, args []string) error {
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
	if err := client.Get(ctx, "/teams/"+teamID+"/memberships", &resp); err != nil {
		return err
	}
	return output.Print(resp)
}
