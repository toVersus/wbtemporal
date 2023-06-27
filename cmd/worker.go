package cmd

import (
	"github.com/spf13/cobra"
)

var (
	workerCmd = &cobra.Command{
		Use:   "worker",
		Short: "Manage Temporal worker to manage Workspace instances",
	}
)