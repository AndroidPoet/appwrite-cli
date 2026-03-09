package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	projectID string
	profile   string
	timeout   string
	dryRun    bool
)

// SetProjectID sets the project ID from the flag
func SetProjectID(p string) {
	projectID = p
}

// SetProfile sets the profile from the flag
func SetProfile(p string) {
	profile = p
}

// SetTimeout sets the timeout from the flag
func SetTimeout(t string) {
	timeout = t
}

// SetDryRun sets the dry-run mode
func SetDryRun(d bool) {
	dryRun = d
}

// GetProjectID returns the project ID from flag, env, or config
func GetProjectID() string {
	if projectID != "" {
		return projectID
	}
	return viper.GetString("project")
}

// GetProfile returns the profile name from flag, env, or config
func GetProfile() string {
	if profile != "" {
		return profile
	}
	p := viper.GetString("profile")
	if p == "" {
		return "default"
	}
	return p
}

// GetTimeout returns the timeout duration string
func GetTimeout() string {
	if timeout != "" {
		return timeout
	}
	t := viper.GetString("timeout")
	if t == "" {
		return "60s"
	}
	return t
}

// IsDryRun returns whether dry-run mode is enabled
func IsDryRun() bool {
	return dryRun
}

// RequireProject validates project ID is set
func RequireProject(cmd *cobra.Command) error {
	pid := GetProjectID()
	if pid == "" {
		return fmt.Errorf("project ID required: use --project flag or set AW_PROJECT environment variable")
	}
	return nil
}

// CheckConfirm validates confirmation for destructive operations
func CheckConfirm(cmd *cobra.Command) error {
	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("this is a destructive operation. Use --confirm to proceed")
	}
	return nil
}
