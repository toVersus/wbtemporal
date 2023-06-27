package cmd

import (
	"github.com/spf13/cobra"
)

var (
	starterCmd = &cobra.Command{
		Use:   "starter",
		Short: "Trigegr Temporal workflow to manage Workspace instances",
	}
)
