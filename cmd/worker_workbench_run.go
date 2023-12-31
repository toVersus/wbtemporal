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
	workerWorkbenchRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run Temporal worker to Vertex AI Workbench instances",
		Run:   workerWorkbenchRun,
	}
)

func workerWorkbenchRun(cmd *cobra.Command, args []string) {
	// Pass to shared google client used by activity worker
	ctx := context.Background()
	logger := logger.NewDefaultLogger(logLevel)

	opts := ExecutorOpts{Name: executorName}
	logger.Info(fmt.Sprintf("executor option: %+v", opts))
	executor, err := NewGoogleAPIExecutor(ctx, opts)
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

	wa := &activity.WorkbenchActivity{
		Executor: executor,
	}

	cw := worker.New(c, workflow.CreateWorkbenchTaskQueue, worker.Options{
		WorkerStopTimeout:         20 * time.Second,
		BackgroundActivityContext: ctx,
	})
	cw.RegisterWorkflow(workflow.CreateWorkbench)
	cw.RegisterActivity(wa)

	dw := worker.New(c, workflow.DeleteWorkbenchTaskQueue, worker.Options{
		WorkerStopTimeout:         20 * time.Second,
		BackgroundActivityContext: ctx,
	})
	dw.RegisterWorkflow(workflow.DeleteWorkbench)
	dw.RegisterActivity(wa)

	tw := worker.New(c, workflow.StartWorkbenchTaskQueue, worker.Options{
		WorkerStopTimeout:         20 * time.Second,
		BackgroundActivityContext: ctx,
	})
	tw.RegisterWorkflow(workflow.StartWorkbench)
	tw.RegisterActivity(wa)

	sw := worker.New(c, workflow.StopWorkbenchTaskQueue, worker.Options{
		WorkerStopTimeout:         20 * time.Second,
		BackgroundActivityContext: ctx,
	})
	sw.RegisterWorkflow(workflow.StopWorkbench)
	sw.RegisterActivity(wa)

	wg := sync.WaitGroup{}
	wg.Add(4)
	go func() {
		if err := cw.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Failed to start create workspace worker: %s", err)
		}
		wg.Done()
	}()
	go func() {
		if err := dw.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Failed to start delete workspace worker: %s", err)
		}
		wg.Done()
	}()

	go func() {
		if err := tw.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Failed to start start workspace worker: %s", err)
		}
		wg.Done()
	}()

	go func() {
		if err := sw.Run(worker.InterruptCh()); err != nil {
			log.Fatalf("Failed to start stop workspace worker: %s", err)
		}
		wg.Done()
	}()

	wg.Wait()
	logger.Info("Successfully stop worker process!")
}
