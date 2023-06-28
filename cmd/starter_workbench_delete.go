package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/executor/googleapi"
	"github.com/toVersus/wbtemporal/pkg/logger"
	"github.com/toVersus/wbtemporal/pkg/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	starterWorkbenchDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Trigger Temporal workflow to delete Workspace instance",
		Run:   starterWorkbenchDelete,
	}
)

func starterWorkbenchDelete(cmd *cobra.Command, args []string) {
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

	options := &googleapi.Option{
		Name:      name,
		Location:  location,
		Zone:      zone,
		ProjectId: projectID,
	}
	workflowID := fmt.Sprintf("%s-delete", name)
	logger.Info("Trigger workflow to delete workspace instance")
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflow.DeleteWorkbenchTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Minute,
			MaximumAttempts: 3,
		},
	}, workflow.DeleteWorkbench, options)
	if err != nil {
		logger.Fatal("Could not trigger delete workspace workflow", "Error", err)
	}
	if !wait {
		logger.Info("Successfully triggered delete workspace workflow!")
		return
	}

	if !silent {
		// Poll and print workflow status using separate goroutine
		watcher := &workflowWatcher{c: c, id: workflowID}
		logger.Info("Start workflow watcher")
		watcher.run(ctx)
	}

	var status googleapi.Status
	if err := run.Get(ctx, &status); err != nil {
		logger.Fatal("Could not complete delete workspace workflow", "Error", err)
	}
	logger.Info("Successfully complte delete workspace workflow!")
	// Just to be sure, sleep 3 seconds before exiting
	time.Sleep(3 * time.Second)
}
