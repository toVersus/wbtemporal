package cmd

import "github.com/spf13/cobra"

var (
	starterWorkbenchCmd = &cobra.Command{
		Use:   "workbench",
		Short: "Trigger Temporal workflow to manage Workspace instance",
	}
)
