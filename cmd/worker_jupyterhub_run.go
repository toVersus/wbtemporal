package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/activity"
	"github.com/toVersus/wbtemporal/pkg/logger"
	"github.com/toVersus/wbtemporal/pkg/workflow"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/worker"
)

var (
	workerJupyterHubRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run Temporal worker to manage JupyterHub instances",
		Run:   workerJupyterHubRun,
	}
)

func workerJupyterHubRun(cmd *cobra.Command, args []string) {
	// Pass to shared google client used by activity worker
	ctx := context.Background()
	logger := logger.NewDefaultLogger(logLevel)

	opts := ExecutorOpts{Name: executorName}
	logger.Info(fmt.Sprintf("executor option: %+v", opts))
	executor, err := NewJupyterHubExecutor(ctx, opts)
	if err != nil {
		logger.Fatal("Failed to select executor: %s", err)
	}

	logger.Debug(fmt.Sprintf("trying to connect to temporal frontend: %s", frontendAddr))
	c, err := client.Dial(client.Options{
		HostPort: fmt.Sprintf("dns:///%s", frontendAddr),
		Logger:   logger,
		MetricsHandler: sdktally.NewMetricsHandler(newPrometheusScope(prometheus.Configuration{
			ListenAddress: "0.0.0.0:9090",
			TimerType:     "histogram",
		})),
	})
	if err != nil {
		logger.Fatal("Failed to create Temporal client", "Error", err)
	}
	defer c.Close()
	logger.Info(fmt.Sprintf("Successfully connected to temporal frontend: %s", frontendAddr))

	wa := &activity.JupyterHubActivity{
		Executor: executor,
	}

	createJupyterHubWorker := worker.New(c, workflow.CreateJupyterHubTaskQueue, worker.Options{
		WorkerStopTimeout:         20 * time.Second,
		BackgroundActivityContext: ctx,
	})
	createJupyterHubWorker.RegisterWorkflow(workflow.CreateUserServer)
	createJupyterHubWorker.RegisterActivity(wa)

	deleteJupyterHubWorker := worker.New(c, workflow.DeleteJupyterHubTaskQueue, worker.Options{
		WorkerStopTimeout:         20 * time.Second,
		BackgroundActivityContext: ctx,
	})
	deleteJupyterHubWorker.RegisterWorkflow(workflow.DeleteUserServer)
	deleteJupyterHubWorker.RegisterActivity(wa)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := createJupyterHubWorker.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Failed to start create JupyterHub user server worker: %s", err)
		}
		wg.Done()
	}()
	go func() {
		if err := deleteJupyterHubWorker.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Failed to start delete JupyterHub user server worker: %s", err)
		}
		wg.Done()
	}()

	wg.Wait()
	logger.Info("Successfully stop JupyterHub worker process!")
}
