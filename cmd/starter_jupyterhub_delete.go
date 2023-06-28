package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/executor/jupyterhubapi"
	"github.com/toVersus/wbtemporal/pkg/logger"
	"github.com/toVersus/wbtemporal/pkg/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	starterJupyterHubDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Trigger Temporal workflow to delete JupyterHub user server",
		Run:   starterJupyterHubDelete,
	}
)

func starterJupyterHubDelete(cmd *cobra.Command, args []string) {
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

	options := &jupyterhubapi.Option{
		Server: jupyterHubServer,
		User:   jupyterHubUser,
	}
	workflowID := fmt.Sprintf("%s-%s-delete", jupyterHubUser, jupyterHubServer)
	logger.Info("Trigger workflow to delete new JupyterHub user server")
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflow.DeleteJupyterHubTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Minute,
			MaximumAttempts: 3,
		},
	}, workflow.DeleteUserServer, options)
	if err != nil {
		logger.Fatal("Could not trigger delete workflow for JupyterHub user server", "Error", err)
	}
	if !wait {
		logger.Info("Successfully triggered delete workflow for JupyterHub user server!")
		return
	}

	if !silent {
		// Poll and print workflow status using separate goroutine
		watcher := &workflowWatcher{c: c, id: workflowID}
		logger.Info("Start workflow watcher")
		watcher.run(ctx)
	}

	var status jupyterhubapi.Status
	if err := run.Get(ctx, &status); err != nil {
		logger.Fatal("Could not complete delete workflow for JupyterHub user server", "Error", err)
	}
	logger.Info("Delete workflow for JupyterHub user server completed successfully", "name", status.Name, "url", status.URL, "status", status.Status)
	// Just to be sure, sleep 3 seconds before exiting
	time.Sleep(3 * time.Second)
}
