package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/executor"
	"github.com/toVersus/wbtemporal/pkg/executor/googleapi"
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

	// worker flags
	executorName string

	rootCmd = &cobra.Command{
		Use:   "wbtemporal",
		Short: "A tool to manage Workspace instances",
	}

	k = koanf.New(".")

	ErrNotFoundExecutor = fmt.Errorf("executor not found")
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(starterCmd)
	workerCmd.AddCommand(workerRunCmd)
	starterCmd.AddCommand(starterCreateCmd)
	starterCmd.AddCommand(starterDeleteCmd)
	starterCmd.AddCommand(starterStartCmd)
	starterCmd.AddCommand(starterStopCmd)

	rootCmd.PersistentFlags().StringVar(&frontendAddr, "frontend-addr", "localhost:7233",
		`temporal frontend addr to connect, use "<host>:<port>" format`)
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", logger.DebugLevel,
		fmt.Sprintf(`set log level for zap logger "%s", "%s", "%s", "%s", "%s"`,
			logger.DebugLevel, logger.InfoLevel, logger.WarnLevel, logger.ErrorLevel, logger.FatalLevel))

	starterCmd.PersistentFlags().StringVar(&name, "name", "", "name of the Workspace instance")
	starterCmd.PersistentFlags().StringVar(&zone, "zone", "asia-northeast1-a", "zone of the Workspace instance")
	starterCmd.PersistentFlags().StringVar(&location, "location", "asia-northeast1", "location of the subnetwork")
	starterCmd.PersistentFlags().StringVar(&projectID, "project-id", "gcp-sample", "Google Cloud project ID")
	starterCmd.PersistentFlags().BoolVar(&wait, "wait", false, "wait for upgrade workflows to be done")
	starterCmd.PersistentFlags().BoolVar(&silent, "silent", false, "silent mode, do not print periodic activity status")
	starterCmd.MarkPersistentFlagRequired("name")

	starterCreateCmd.Flags().StringVar(&email, "email", "", "Google account email address")
	starterCreateCmd.Flags().StringVar(&machineType, "machine-type", "n1-standard-1", "machine type of the Workspace instance")
	starterCreateCmd.Flags().StringVar(&network, "network", "", "VPC network name that Workspace instance belongs to")
	starterCreateCmd.Flags().StringVar(&subnet, "subnet", "", "VPC subnet name that Workspace instance belongs to")
	starterCreateCmd.MarkPersistentFlagRequired("email")
	starterCreateCmd.MarkPersistentFlagRequired("network")
	starterCreateCmd.MarkPersistentFlagRequired("subnet")

	workerRunCmd.Flags().StringVar(&executorName, "executor-name", executor.ExecutorNameGoogleAPI,
		fmt.Sprintf(`change backend implementation to intract with Google Cloud, current available executor is "%s" and "%s"`,
			executor.ExecutorNameGoogleAPI, executor.ExecutorNameFakeClient))
	logger := logger.NewDefaultLogger(logLevel)

	if len(strings.Split(frontendAddr, ":")) != 2 {
		logger.Fatal(`Invalid format found in frontend addr, use "<host>:<port>" format instead`, "FrontendAddr", frontendAddr)
		os.Exit(1)
	}
}

type ExecutorOpts struct {
	Name string
}

func NewExecutor(ctx context.Context, opts ExecutorOpts) (executor.Executor, error) {
	if opts.Name == executor.ExecutorNameGoogleAPI {
		return googleapi.NewWorkbench(ctx)
	}
	// } else if opts.Name == executor.ExecutorNameFakeClient {
	// 	return fakeclient.NewFakeClientExecutor(), nil
	// }
	return nil, fmt.Errorf("executor %s not supported: %w", opts.Name, ErrNotFoundExecutor)
}
