package users

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
	userID   string
	email    string
	password string
	name     string
	phone    string
	limit    int
	offset   int
	allPages bool
)

// UsersCmd manages users
var UsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  `List, create, update, and delete users.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE:  runList,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get user details",
	RunE:  runGet,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	RunE:  runCreate,
}

var updateNameCmd = &cobra.Command{
	Use:   "update-name",
	Short: "Update user name",
	RunE:  runUpdateName,
}

var updateEmailCmd = &cobra.Command{
	Use:   "update-email",
	Short: "Update user email",
	RunE:  runUpdateEmail,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	RunE:  runDelete,
}

var listSessionsCmd = &cobra.Command{
	Use:   "list-sessions",
	Short: "List user sessions",
	RunE:  runListSessions,
}

var listLogsCmd = &cobra.Command{
	Use:   "list-logs",
	Short: "List user logs",
	RunE:  runListLogs,
}

func init() {
	listCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getCmd.Flags().StringVar(&userID, "user-id", "", "user ID")
	getCmd.MarkFlagRequired("user-id")
	getCmd.RegisterFlagCompletionFunc("user-id", completion.UserIDs())

	createCmd.Flags().StringVar(&userID, "user-id", "", "user ID (auto-generated if empty)")
	createCmd.Flags().StringVar(&email, "email", "", "user email")
	createCmd.Flags().StringVar(&password, "password", "", "user password")
	createCmd.Flags().StringVar(&name, "name", "", "user name")
	createCmd.Flags().StringVar(&phone, "phone", "", "user phone")
	createCmd.MarkFlagRequired("email")
	createCmd.MarkFlagRequired("password")

	updateNameCmd.Flags().StringVar(&userID, "user-id", "", "user ID")
	updateNameCmd.Flags().StringVar(&name, "name", "", "new name")
	updateNameCmd.MarkFlagRequired("user-id")
	updateNameCmd.MarkFlagRequired("name")
	updateNameCmd.RegisterFlagCompletionFunc("user-id", completion.UserIDs())

	updateEmailCmd.Flags().StringVar(&userID, "user-id", "", "user ID")
	updateEmailCmd.Flags().StringVar(&email, "email", "", "new email")
	updateEmailCmd.MarkFlagRequired("user-id")
	updateEmailCmd.MarkFlagRequired("email")
	updateEmailCmd.RegisterFlagCompletionFunc("user-id", completion.UserIDs())

	var confirm bool
	deleteCmd.Flags().StringVar(&userID, "user-id", "", "user ID")
	deleteCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteCmd.MarkFlagRequired("user-id")
	deleteCmd.RegisterFlagCompletionFunc("user-id", completion.UserIDs())

	listSessionsCmd.Flags().StringVar(&userID, "user-id", "", "user ID")
	listSessionsCmd.MarkFlagRequired("user-id")
	listSessionsCmd.RegisterFlagCompletionFunc("user-id", completion.UserIDs())

	listLogsCmd.Flags().StringVar(&userID, "user-id", "", "user ID")
	listLogsCmd.MarkFlagRequired("user-id")
	listLogsCmd.RegisterFlagCompletionFunc("user-id", completion.UserIDs())

	UsersCmd.AddCommand(listCmd)
	UsersCmd.AddCommand(getCmd)
	UsersCmd.AddCommand(createCmd)
	UsersCmd.AddCommand(updateNameCmd)
	UsersCmd.AddCommand(updateEmailCmd)
	UsersCmd.AddCommand(deleteCmd)
	UsersCmd.AddCommand(listSessionsCmd)
	UsersCmd.AddCommand(listLogsCmd)
}

type UserInfo struct {
	ID            string `json:"$id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone,omitempty"`
	Status        bool   `json:"status"`
	EmailVerified bool   `json:"emailVerification"`
	CreatedAt     string `json:"$createdAt,omitempty"`
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
		var items []UserInfo
		err := client.ListAll(ctx, "/users", limit, "users", func(raw json.RawMessage) error {
			var page []UserInfo
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
		Users []UserInfo `json:"users"`
	}
	path := fmt.Sprintf("/users?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Users)
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

	var user UserInfo
	if err := client.Get(ctx, "/users/"+userID, &user); err != nil {
		return err
	}
	return output.Print(user)
}

func runCreate(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create user '%s'", email)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	if userID != "" {
		body["userId"] = userID
	} else {
		body["userId"] = "unique()"
	}
	if name != "" {
		body["name"] = name
	}
	if phone != "" {
		body["phone"] = phone
	}

	var user UserInfo
	if err := client.Post(ctx, "/users", body, &user); err != nil {
		return err
	}

	output.PrintSuccess("User '%s' created (ID: %s)", email, user.ID)
	return output.Print(user)
}

func runUpdateName(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	var user UserInfo
	if err := client.Patch(ctx, "/users/"+userID+"/name", map[string]interface{}{"name": name}, &user); err != nil {
		return err
	}

	output.PrintSuccess("User '%s' name updated", userID)
	return output.Print(user)
}

func runUpdateEmail(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	var user UserInfo
	if err := client.Patch(ctx, "/users/"+userID+"/email", map[string]interface{}{"email": email}, &user); err != nil {
		return err
	}

	output.PrintSuccess("User '%s' email updated", userID)
	return output.Print(user)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete user '%s'", userID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, "/users/"+userID); err != nil {
		return err
	}

	output.PrintSuccess("User '%s' deleted", userID)
	return nil
}

func runListSessions(cmd *cobra.Command, args []string) error {
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
	if err := client.Get(ctx, "/users/"+userID+"/sessions", &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runListLogs(cmd *cobra.Command, args []string) error {
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
	if err := client.Get(ctx, "/users/"+userID+"/logs", &resp); err != nil {
		return err
	}
	return output.Print(resp)
}
