package cmd

import "github.com/spf13/cobra"

var (
	starterJupyterHubCmd = &cobra.Command{
		Use:   "jupyterhub",
		Short: "Trigger Temporal workflow to manage JupyterHub user server",
	}
)
