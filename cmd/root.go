package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/executor/googleapi"
	"github.com/toVersus/wbtemporal/pkg/executor/jupyterhubapi"
	"github.com/toVersus/wbtemporal/pkg/logger"
)

var (
	// global flags
	frontendAddr string
	logLevel     string

	// starter flags
	name        string
	zone        string
	location    string
	projectID   string
	email       string
	machineType string
	network     string
	subnet      string
	wait        bool
	silent      bool

	jupyterHubUser   string
	jupyterHubServer string

	// worker flags
	executorName string

	jupyterHubBaseURL  string
	jupyterHubAPIToken string

	rootCmd = &cobra.Command{
		Use:   "wbtemporal",
		Short: "A tool to manage Workspace instances",
	}

	ErrNotFoundExecutor = fmt.Errorf("executor not found")
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(starterCmd)

	starterCmd.AddCommand(starterWorkbenchCmd)
	starterCmd.AddCommand(starterJupyterHubCmd)
	workerCmd.AddCommand(workerWorkbenchCmd)
	workerCmd.AddCommand(workerJupyterHubCmd)

	workerWorkbenchCmd.AddCommand(workerWorkbenchRunCmd)
	workerJupyterHubCmd.AddCommand(workerJupyterHubRunCmd)

	starterJupyterHubCmd.AddCommand(starterJupyterHubCreateCmd)
	starterJupyterHubCmd.AddCommand(starterJupyterHubDeleteCmd)

	starterWorkbenchCmd.AddCommand(starterWorkbenchCreateCmd)
	starterWorkbenchCmd.AddCommand(starterWorkbenchDeleteCmd)
	starterWorkbenchCmd.AddCommand(starterWorkbenchStartCmd)
	starterWorkbenchCmd.AddCommand(starterWorkbenchStopCmd)

	rootCmd.PersistentFlags().StringVar(&frontendAddr, "frontend-addr", "localhost:7233",
		`temporal frontend addr to connect, use "<host>:<port>" format`)
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", logger.DebugLevel,
		fmt.Sprintf(`set log level for zap logger "%s", "%s", "%s", "%s", "%s"`,
			logger.DebugLevel, logger.InfoLevel, logger.WarnLevel, logger.ErrorLevel, logger.FatalLevel))

	starterCmd.PersistentFlags().BoolVar(&wait, "wait", false, "wait for upgrade workflows to be done")

	starterWorkbenchCmd.PersistentFlags().StringVar(&name, "name", "", "name of the Workspace instance")
	starterWorkbenchCmd.PersistentFlags().StringVar(&zone, "zone", "asia-northeast1-a", "zone of the Workspace instance")
	starterWorkbenchCmd.PersistentFlags().StringVar(&location, "location", "asia-northeast1", "location of the subnetwork")
	starterWorkbenchCmd.PersistentFlags().StringVar(&projectID, "project-id", "gcp-sample", "Google Cloud project ID")
	starterWorkbenchCmd.PersistentFlags().BoolVar(&silent, "silent", false, "silent mode, do not print periodic activity status")
	starterWorkbenchCmd.MarkPersistentFlagRequired("name")

	starterJupyterHubCmd.PersistentFlags().StringVar(&jupyterHubUser, "user", "", "JupyterHub user name")
	starterJupyterHubCmd.PersistentFlags().StringVar(&jupyterHubServer, "server", "", "JupyterHub user server name")

	starterWorkbenchCreateCmd.Flags().StringVar(&email, "email", "", "Google account email address")
	starterWorkbenchCreateCmd.Flags().StringVar(&machineType, "machine-type", "n1-standard-1", "machine type of the Workspace instance")
	starterWorkbenchCreateCmd.Flags().StringVar(&network, "network", "", "VPC network name that Workspace instance belongs to")
	starterWorkbenchCreateCmd.Flags().StringVar(&subnet, "subnet", "", "VPC subnet name that Workspace instance belongs to")
	starterWorkbenchCreateCmd.MarkPersistentFlagRequired("email")
	starterWorkbenchCreateCmd.MarkPersistentFlagRequired("network")
	starterWorkbenchCreateCmd.MarkPersistentFlagRequired("subnet")

	workerWorkbenchRunCmd.Flags().StringVar(&executorName, "executor-name", googleapi.ExecutorNameGoogleAPI,
		fmt.Sprintf(`change backend implementation to intract with Google Cloud, current available executor is %q and %q for testing`,
			googleapi.ExecutorNameGoogleAPI, googleapi.ExecutorNameFakeClient))

	workerJupyterHubRunCmd.Flags().StringVar(&jupyterHubBaseURL, "base-url", "", "JupyterHub base URL")
	workerJupyterHubRunCmd.Flags().StringVar(&jupyterHubAPIToken, "token", "", "JupyterHub API token")
	workerJupyterHubRunCmd.Flags().StringVar(&executorName, "executor-name", googleapi.ExecutorNameGoogleAPI,
		fmt.Sprintf(`change backend implementation to intract with Google Cloud, current available executor is %q and %q for testing`,
			jupyterhubapi.ExecutorNameJupyterHub, googleapi.ExecutorNameFakeClient))

	logger := logger.NewDefaultLogger(logLevel)

	if len(strings.Split(frontendAddr, ":")) != 2 {
		logger.Fatal(`Invalid format found in frontend addr, use "<host>:<port>" format instead`, "FrontendAddr", frontendAddr)
		os.Exit(1)
	}
}

type ExecutorOpts struct {
	Name string
}
