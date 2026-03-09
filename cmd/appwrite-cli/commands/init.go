package commands

import (
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/auth"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/collections"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/completion"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/databases"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/diff"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/doctor"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/documents"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/exportcmd"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/functions"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/health"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/initcmd"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/messaging"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/report"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/status"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/storage"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/teams"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/users"
	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands/watch"
)

func init() {
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(databases.DatabasesCmd)
	rootCmd.AddCommand(collections.CollectionsCmd)
	rootCmd.AddCommand(documents.DocumentsCmd)
	rootCmd.AddCommand(storage.StorageCmd)
	rootCmd.AddCommand(functions.FunctionsCmd)
	rootCmd.AddCommand(users.UsersCmd)
	rootCmd.AddCommand(teams.TeamsCmd)
	rootCmd.AddCommand(messaging.MessagingCmd)
	rootCmd.AddCommand(health.HealthCmd)
	rootCmd.AddCommand(doctor.DoctorCmd)
	rootCmd.AddCommand(status.StatusCmd)
	rootCmd.AddCommand(watch.WatchCmd)
	rootCmd.AddCommand(diff.DiffCmd)
	rootCmd.AddCommand(exportcmd.ExportCmd)
	rootCmd.AddCommand(exportcmd.ImportCmd)
	rootCmd.AddCommand(report.ReportCmd)
	rootCmd.AddCommand(completion.CompletionCmd)
	rootCmd.AddCommand(initcmd.InitCmd)
}
