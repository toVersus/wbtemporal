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
	starterWorkbenchCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Trigger Temporal workflow to create Workspace instance",
		Run:   starterWorkbenchCreate,
	}
)

func starterWorkbenchCreate(cmd *cobra.Command, args []string) {
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
		Name:        name,
		Email:       email,
		Location:    location,
		Zone:        zone,
		ProjectId:   projectID,
		MachineType: machineType,
		Network:     network,
		Subnet:      subnet,
	}
	workflowID := fmt.Sprintf("%s-create", name)
	logger.Info("Trigger workflow to create new workspace instance")
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflow.CreateWorkbenchTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Minute,
			MaximumAttempts: 3,
		},
	}, workflow.CreateWorkbench, options)
	if err != nil {
		logger.Fatal("Could not trigger create workspace workflow", "Error", err)
	}
	if !wait {
		logger.Info("Successfully triggered create workspace workflow!")
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
		logger.Fatal("Could not complete create workspace workflow", "Error", err)
	}
	logger.Info("Workspace workflow completed successfully", "name", status.Name, "url", status.URL, "status", status.Status)
	// Just to be sure, sleep 3 seconds before exiting
	time.Sleep(3 * time.Second)
}
