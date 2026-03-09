package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/initcmd"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/config"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

var (
	cfgFile   string
	projectID string
	profile   string
	endpoint  string
	outputFmt string
	pretty    bool
	quiet     bool
	debug     bool
	timeout   string
	dryRun    bool

	versionStr string
	commitStr  string
	dateStr    string
)

var rootCmd = &cobra.Command{
	Use:   "appwrite-cli",
	Short: "Appwrite CLI",
	Long: `appwrite-cli is a fast, single-binary CLI for Appwrite.

Built for power users and CI/CD — multi-profile, multi-format,
multi-environment. Manage databases, storage, functions, users,
and more from the terminal.

Design Philosophy:
  • JSON-first output for automation
  • Multi-profile auth (cloud + self-hosted)
  • Six output formats: json, table, csv, tsv, yaml, minimal
  • No interactive prompts — fully scriptable
  • Clean exit codes (0=success, 1=error, 2=validation)`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "completion" || cmd.Name() == "help" || cmd.Name() == "__complete" {
			return nil
		}

		// Load .aw.yaml project config if it exists
		if cwd, err := os.Getwd(); err == nil {
			if projectCfg := initcmd.FindProjectConfig(cwd); projectCfg != "" {
				viper.SetConfigFile(projectCfg)
				viper.SetConfigType("yaml")
				_ = viper.MergeInConfig()
			}
		}

		// Sync flags to cli package
		cli.SetProjectID(projectID)
		cli.SetProfile(profile)
		cli.SetTimeout(timeout)
		cli.SetDryRun(dryRun)

		// Initialize config
		if err := config.Init(cfgFile, profile); err != nil {
			return err
		}

		// Setup output formatter
		output.Setup(outputFmt, pretty, quiet)

		if debug {
			config.SetDebug(true)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets the version information
func SetVersionInfo(version, commit, date string) {
	versionStr = version
	commitStr = commit
	dateStr = date
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.appwrite-cli/config.json)")
	rootCmd.PersistentFlags().StringVarP(&projectID, "project", "p", "", "Appwrite project ID (or AW_PROJECT env)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "auth profile name (or AW_PROFILE env)")
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "Appwrite endpoint URL (or AW_ENDPOINT env)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "json", "output format: json, table, minimal, tsv, csv, yaml")
	rootCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty-print JSON output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "show API requests/responses")
	rootCmd.PersistentFlags().StringVar(&timeout, "timeout", "60s", "request timeout")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview changes without applying")

	// Bind to viper
	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("endpoint", rootCmd.PersistentFlags().Lookup("endpoint"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

	// Environment variable bindings
	viper.BindEnv("project", "AW_PROJECT")
	viper.BindEnv("profile", "AW_PROFILE")
	viper.BindEnv("endpoint", "AW_ENDPOINT")
	viper.BindEnv("output", "AW_OUTPUT")
	viper.BindEnv("debug", "AW_DEBUG")
	viper.BindEnv("timeout", "AW_TIMEOUT")
	viper.BindEnv("api_key", "AW_API_KEY")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("appwrite-cli %s\n", versionStr)
			fmt.Printf("  commit: %s\n", commitStr)
			fmt.Printf("  built:  %s\n", dateStr)
		},
	})
}

// GetRootCmd returns the root command for adding subcommands
func GetRootCmd() *cobra.Command {
	return rootCmd
}
