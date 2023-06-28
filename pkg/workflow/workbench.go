package workflow

import (
	"fmt"
	"time"

	"github.com/toVersus/wbtemporal/pkg/activity"
	"github.com/toVersus/wbtemporal/pkg/executor/googleapi"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	CreateWorkbenchTaskQueue = "CREATE_WORKBENCH_TASK_QUEUE"
	DeleteWorkbenchTaskQueue = "DELETE_WORKBENCH_TASK_QUEUE"
	StartWorkbenchTaskQueue  = "START_WORKBENCH_TASK_QUEUE"
	StopWorkbenchTaskQueue   = "STOP_WORKBENCH_TASK_QUEUE"
)

func CreateWorkbench(ctx workflow.Context, option *googleapi.Option) (*googleapi.Status, error) {
	var wa *activity.WorkbenchActivity

	logger := defaultGoogleAPIWorkflowLogger(ctx, option)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// アクティビティの実行時間のタイムアウト値
		StartToCloseTimeout: 1 * time.Minute,
		// アクティビティを 5 秒間隔で 72 回の合計 6 分間リトライする
		// Vertex AI Workbench のインスタンスの作成を待つ時のリトライ戦略をベースに設定
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        5 * time.Second,
			MaximumInterval:        5 * time.Second,
			MaximumAttempts:        72,
			NonRetryableErrorTypes: []string{activity.ErrLongRunningOperationFailed},
		},
	})

	logger.Info("Checking for the existence of Workbench instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		return nil, fmt.Errorf("failed to check for the existence of Workbench instance: %w", err)
	}

	if exist {
		logger.Info("Workbench instance already exists")
	} else {
		logger.Info("Creating new Workbench instance")
		var opName string
		if err := workflow.ExecuteActivity(ctx, wa.Create, option).Get(ctx, &opName); err != nil {
			return nil, fmt.Errorf("failed to create Workbench instance: %w", err)
		}

		logger.Info("Waiting for Workbench instance created")
		if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("failed to watch operation for creation of Workbench instance: %w", err)
		}
	}

	var status googleapi.Status
	logger.Info("Waiting for instance to be provisioned and getting URL for accessing to Workbench instance")
	if err := workflow.ExecuteActivity(ctx, wa.GetWorkspaceURL, option).Get(ctx, &status); err != nil {
		return nil, fmt.Errorf("failed to watch operation to create Workbench instance: %w", err)
	}

	logger.Info("Workbench instance created successfully!")
	return &status, nil
}

func DeleteWorkbench(ctx workflow.Context, option *googleapi.Option) error {
	var wa *activity.WorkbenchActivity

	logger := defaultGoogleAPIWorkflowLogger(ctx, option)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// アクティビティの実行時間のタイムアウト値
		StartToCloseTimeout: 1 * time.Minute,
		// アクティビティを 5 秒間隔で 36 回の合計 3 分間リトライする
		// Vertex AI Workbench のインスタンスの削除を待つ時のリトライ戦略をベースに設定
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        5 * time.Second,
			MaximumInterval:        5 * time.Second,
			MaximumAttempts:        36,
			NonRetryableErrorTypes: []string{activity.ErrLongRunningOperationFailed},
		},
	})

	logger.Info("Checking for the existence of Workbench instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		logger.Info("Workbench instance already not exists")
		return nil
	}
	if !exist {
		logger.Info("Workbench instance already deleted")
	}

	logger.Info("Deleting Workbench instance")
	var opName string
	if err := workflow.ExecuteActivity(ctx, wa.Delete, option).Get(ctx, &opName); err != nil {
		return fmt.Errorf("failed to delete Workbench instance: %w", err)
	}

	logger.Info("Waiting for Workbench instance deleted")
	if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to watch operation to delete Workbench instance: %w", err)
	}

	logger.Info("Workbench instance deleted successfully!")
	return nil
}

func StartWorkbench(ctx workflow.Context, option *googleapi.Option) (*googleapi.Status, error) {
	var wa *activity.WorkbenchActivity

	logger := defaultGoogleAPIWorkflowLogger(ctx, option)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// アクティビティの実行時間のタイムアウト値
		StartToCloseTimeout: 1 * time.Minute,
		// アクティビティを 5 秒間隔で 72 回の合計 6 分間リトライする
		// Vertex AI Workbench のインスタンスの起動を待つ時のリトライ戦略をベースに設定
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        5 * time.Second,
			MaximumInterval:        5 * time.Second,
			MaximumAttempts:        72,
			NonRetryableErrorTypes: []string{activity.ErrLongRunningOperationFailed},
		},
	})

	logger.Info("Checking for the existence of Workbench instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		return nil, fmt.Errorf("workbench instance not found: %w", err)
	}

	logger.Info("Starting Workbench instance")
	var opName string
	if err := workflow.ExecuteActivity(ctx, wa.Start, option).Get(ctx, &opName); err != nil {
		return nil, fmt.Errorf("failed to start Workbench instance: %w", err)
	}

	logger.Info("Waiting for Workbench instance started")
	if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to watch operation to start Workbench instance: %w", err)
	}

	logger.Info("Getting URL for accessing to Workbench")
	var status googleapi.Status
	if err := workflow.ExecuteActivity(ctx, wa.GetWorkspaceURL, option).Get(ctx, &status); err != nil {
		return nil, fmt.Errorf("failed to watch operation to create Workbench instance: %w", err)
	}

	logger.Info("Workbench instance started successfully!")
	return &status, nil
}

func StopWorkbench(ctx workflow.Context, option *googleapi.Option) error {
	var wa *activity.WorkbenchActivity

	logger := defaultGoogleAPIWorkflowLogger(ctx, option)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// アクティビティの実行時間のタイムアウト値
		StartToCloseTimeout: 1 * time.Minute,
		// アクティビティを 5 秒間隔で 36 回の合計 3 分間リトライする
		// Vertex AI Workbench のインスタンスの停止を待つ時のリトライ戦略をベースに設定
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        5 * time.Second,
			MaximumInterval:        5 * time.Second,
			MaximumAttempts:        72,
			NonRetryableErrorTypes: []string{activity.ErrLongRunningOperationFailed},
		},
	})

	logger.Info("Checking for the existence of Workbench instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		return fmt.Errorf("workbench instance not found: %w", err)
	}

	logger.Info("Stopping Workbench instance")
	var opName string
	if err := workflow.ExecuteActivity(ctx, wa.Stop, option).Get(ctx, &opName); err != nil {
		return fmt.Errorf("failed to stop Workbench instance: %w", err)
	}

	logger.Info("Waiting for Workbench instance stopped")
	if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to watch operation to stop Workbench instance: %w", err)
	}

	logger.Info("Workbench instance stopped successfully!")
	return nil
}
