package completion

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
)

// DatabaseIDs returns a completion function for database IDs
func DatabaseIDs() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		items := fetchIDs("/databases", "databases", "$id", "name")
		return items, cobra.ShellCompDirectiveNoFileComp
	}
}

// CollectionIDs returns a completion function for collection IDs
func CollectionIDs() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// BucketIDs returns a completion function for bucket IDs
func BucketIDs() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		items := fetchIDs("/storage/buckets", "buckets", "$id", "name")
		return items, cobra.ShellCompDirectiveNoFileComp
	}
}

// FunctionIDs returns a completion function for function IDs
func FunctionIDs() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		items := fetchIDs("/functions", "functions", "$id", "name")
		return items, cobra.ShellCompDirectiveNoFileComp
	}
}

// TeamIDs returns a completion function for team IDs
func TeamIDs() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		items := fetchIDs("/teams", "teams", "$id", "name")
		return items, cobra.ShellCompDirectiveNoFileComp
	}
}

// UserIDs returns a completion function for user IDs
func UserIDs() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		items := fetchIDs("/users", "users", "$id", "name")
		return items, cobra.ShellCompDirectiveNoFileComp
	}
}

func fetchIDs(path, itemsKey, idField, descField string) []string {
	projectID := cli.GetProjectID()
	if projectID == "" {
		return nil
	}

	client, err := api.NewClient(projectID, 5*time.Second)
	if err != nil {
		return nil
	}

	ctx, cancel := client.Context()
	defer cancel()

	var resp map[string]interface{}
	if err := client.Get(ctx, path+"?limit=20", &resp); err != nil {
		return nil
	}

	items, ok := resp[itemsKey].([]interface{})
	if !ok {
		return nil
	}

	var results []string
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id, ok := m[idField].(string)
		if !ok {
			continue
		}
		desc, _ := m[descField].(string)
		if desc != "" {
			results = append(results, fmt.Sprintf("%s\t%s", id, desc))
		} else {
			results = append(results, id)
		}
	}
	return results
}
