package messaging

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

var (
	messageID string
	topicID   string
	subject   string
	content   string
	targets   []string
	limit     int
	offset    int
)

// MessagingCmd manages messaging
var MessagingCmd = &cobra.Command{
	Use:   "messaging",
	Short: "Manage messaging",
	Long:  `Manage messages, topics, subscribers, and providers.`,
}

var listMessagesCmd = &cobra.Command{
	Use:   "list-messages",
	Short: "List all messages",
	RunE:  runListMessages,
}

var getMessageCmd = &cobra.Command{
	Use:   "get-message",
	Short: "Get message details",
	RunE:  runGetMessage,
}

var listTopicsCmd = &cobra.Command{
	Use:   "list-topics",
	Short: "List all topics",
	RunE:  runListTopics,
}

var getTopicCmd = &cobra.Command{
	Use:   "get-topic",
	Short: "Get topic details",
	RunE:  runGetTopic,
}

var createTopicCmd = &cobra.Command{
	Use:   "create-topic",
	Short: "Create a new topic",
	RunE:  runCreateTopic,
}

var deleteTopicCmd = &cobra.Command{
	Use:   "delete-topic",
	Short: "Delete a topic",
	RunE:  runDeleteTopic,
}

var listProvidersCmd = &cobra.Command{
	Use:   "list-providers",
	Short: "List messaging providers",
	RunE:  runListProviders,
}

func init() {
	listMessagesCmd.Flags().IntVar(&limit, "limit", 25, "number of results")
	listMessagesCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")

	getMessageCmd.Flags().StringVar(&messageID, "message-id", "", "message ID")
	getMessageCmd.MarkFlagRequired("message-id")

	listTopicsCmd.Flags().IntVar(&limit, "limit", 25, "number of results")
	listTopicsCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")

	getTopicCmd.Flags().StringVar(&topicID, "topic-id", "", "topic ID")
	getTopicCmd.MarkFlagRequired("topic-id")

	createTopicCmd.Flags().StringVar(&topicID, "topic-id", "", "topic ID (auto-generated if empty)")
	createTopicCmd.Flags().StringVar(&subject, "name", "", "topic name")
	createTopicCmd.MarkFlagRequired("name")

	var confirm bool
	deleteTopicCmd.Flags().StringVar(&topicID, "topic-id", "", "topic ID")
	deleteTopicCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteTopicCmd.MarkFlagRequired("topic-id")

	listProvidersCmd.Flags().IntVar(&limit, "limit", 25, "number of results")
	listProvidersCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")

	MessagingCmd.AddCommand(listMessagesCmd)
	MessagingCmd.AddCommand(getMessageCmd)
	MessagingCmd.AddCommand(listTopicsCmd)
	MessagingCmd.AddCommand(getTopicCmd)
	MessagingCmd.AddCommand(createTopicCmd)
	MessagingCmd.AddCommand(deleteTopicCmd)
	MessagingCmd.AddCommand(listProvidersCmd)
}

func parseTimeout() time.Duration {
	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		return 60 * time.Second
	}
	return d
}

func runListMessages(cmd *cobra.Command, args []string) error {
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
	path := fmt.Sprintf("/messaging/messages?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runGetMessage(cmd *cobra.Command, args []string) error {
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
	if err := client.Get(ctx, "/messaging/messages/"+messageID, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runListTopics(cmd *cobra.Command, args []string) error {
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
	path := fmt.Sprintf("/messaging/topics?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runGetTopic(cmd *cobra.Command, args []string) error {
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
	if err := client.Get(ctx, "/messaging/topics/"+topicID, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}

func runCreateTopic(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create topic '%s'", subject)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name": subject,
	}
	if topicID != "" {
		body["topicId"] = topicID
	} else {
		body["topicId"] = "unique()"
	}

	var resp map[string]interface{}
	if err := client.Post(ctx, "/messaging/topics", body, &resp); err != nil {
		return err
	}

	output.PrintSuccess("Topic created")
	return output.Print(resp)
}

func runDeleteTopic(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, "/messaging/topics/"+topicID); err != nil {
		return err
	}

	output.PrintSuccess("Topic '%s' deleted", topicID)
	return nil
}

func runListProviders(cmd *cobra.Command, args []string) error {
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
	path := fmt.Sprintf("/messaging/providers?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}
	return output.Print(resp)
}
