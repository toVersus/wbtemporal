package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/executor"
	"github.com/toVersus/wbtemporal/pkg/logger"
	"github.com/toVersus/wbtemporal/pkg/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	starterStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Trigger Temporal workflow to stop Workspace instance",
		Run:   starterStop,
	}
)

func starterStop(cmd *cobra.Command, args []string) {
	logger := logger.NewDefaultLogger(logLevel)

	logger.Debug(fmt.Sprintf("Trying to connect to temporal frontend: %s", frontendAddr))
	c, err := client.Dial(client.Options{
		HostPort: fmt.Sprintf("dns:///%s", frontendAddr),
		Logger:   logger,
	})
	if err != nil {
		logger.Fatal("Failed to create Temporal client", "Error", err)
	}
	defer c.Close()
	logger.Info(fmt.Sprintf("Successfully connected to temporal frontend: %s", frontendAddr))

	logger.Info("Register signal handler to shutdown starter process gracefully")
	ctx, shutdown := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdown()

	options := &executor.WorkspaceOption{
		Name: name,
		GoogleAPIOption: &executor.GoogleAPIOption{
			Location:  location,
			Zone:      zone,
			ProjectId: projectID,
		},
	}
	workflowID := fmt.Sprintf("%s-stop", name)
	logger.Info("Trigger workflow to stop workspace instance")
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflow.StopWorkspaceTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Minute,
			MaximumAttempts: 3,
		},
	}, workflow.StopWorkspace, options)
	if err != nil {
		logger.Fatal("Could not trigger stop workspace workflow", "Error", err)
	}
	if !wait {
		logger.Info("Successfully triggered stop workspace workflow!")
		return
	}

	if !silent {
		// Poll and print workflow status using separate goroutine
		watcher := &workflowWatcher{c: c, id: workflowID}
		logger.Info("Start workflow watcher")
		watcher.run(ctx, options)
	}

	var status executor.WorkspaceStatus
	if err := run.Get(ctx, &status); err != nil {
		logger.Fatal("Could not complete stop workspace workflow", "Error", err)
	}
	logger.Info("Successfully complte stop workspace workflow!")
	// Just to be sure, sleep 3 seconds before exiting
	time.Sleep(3 * time.Second)
}
