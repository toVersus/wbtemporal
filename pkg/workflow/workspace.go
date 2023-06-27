package workflow

import (
	"fmt"
	"time"

	"github.com/toVersus/wbtemporal/pkg/activity"
	"github.com/toVersus/wbtemporal/pkg/executor"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	CreateWorkspaceTaskQueue = "CREATE_WORKSPACE_TASK_QUEUE"
	DeleteWorkspaceTaskQueue = "DELETE_WORKSPACE_TASK_QUEUE"
	StartWorkspaceTaskQueue  = "START_WORKSPACE_TASK_QUEUE"
	StopWorkspaceTaskQueue   = "STOP_WORKSPACE_TASK_QUEUE"
)

func CreateWorkspace(ctx workflow.Context, option *executor.WorkspaceOption) (*executor.WorkspaceStatus, error) {
	var wa *activity.WorkspaceActvity

	logger := defaultWorkflowLogger(ctx, option)

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

	logger.Info("Checking for the existence of Workspace instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		return nil, fmt.Errorf("failed to check for the existence of Workspace instance: %w", err)
	}

	if exist {
		logger.Info("Workspace instance already exists")
	} else {
		logger.Info("Creating new workspace instance")
		var opName string
		if err := workflow.ExecuteActivity(ctx, wa.Create, option).Get(ctx, &opName); err != nil {
			return nil, fmt.Errorf("failed to create workspace instance: %w", err)
		}

		logger.Info("Waiting for workspace instance created")
		if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("failed to watch operation for creation of workspace instance: %w", err)
		}
	}

	var status executor.WorkspaceStatus
	logger.Info("Waiting for instance to be provisioned and getting URL for accessing to workspace")
	if err := workflow.ExecuteActivity(ctx, wa.GetWorkspaceURL, option).Get(ctx, &status); err != nil {
		return nil, fmt.Errorf("failed to watch operation to create workspace instance: %w", err)
	}

	logger.Info("Workspace instance created successfully!")
	return &status, nil
}

func DeleteWorkspace(ctx workflow.Context, option *executor.WorkspaceOption) error {
	var wa *activity.WorkspaceActvity

	logger := defaultWorkflowLogger(ctx, option)

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

	logger.Info("Checking for the existence of Workspace instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		logger.Info("Workspace instance already not exists")
		return nil
	}
	if !exist {
		logger.Info("Workspace instance already deleted")
	}

	logger.Info("Deleting Workspace instance")
	var opName string
	if err := workflow.ExecuteActivity(ctx, wa.Delete, option).Get(ctx, &opName); err != nil {
		return fmt.Errorf("failed to delete Workspace instance: %w", err)
	}

	logger.Info("Waiting for Workspace instance deleted")
	if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to watch operation to delete Workspace instance: %w", err)
	}

	logger.Info("Workspace instance deleted successfully!")
	return nil
}

func StartWorkspace(ctx workflow.Context, option *executor.WorkspaceOption) (*executor.WorkspaceStatus, error) {
	var wa *activity.WorkspaceActvity

	logger := defaultWorkflowLogger(ctx, option)

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

	logger.Info("Checking for the existence of Workspace instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		return nil, fmt.Errorf("workspace instance not found: %w", err)
	}

	logger.Info("Starting Workspace instance")
	var opName string
	if err := workflow.ExecuteActivity(ctx, wa.Start, option).Get(ctx, &opName); err != nil {
		return nil, fmt.Errorf("failed to start Workspace instance: %w", err)
	}

	logger.Info("Waiting for Workspace instance started")
	if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to watch operation to start Workspace instance: %w", err)
	}

	logger.Info("Getting URL for accessing to workspace")
	var status executor.WorkspaceStatus
	if err := workflow.ExecuteActivity(ctx, wa.GetWorkspaceURL, option).Get(ctx, &status); err != nil {
		return nil, fmt.Errorf("failed to watch operation to create workspace instance: %w", err)
	}

	logger.Info("Workspace instance started successfully!")
	return &status, nil
}

func StopWorkspace(ctx workflow.Context, option *executor.WorkspaceOption) error {
	var wa *activity.WorkspaceActvity

	logger := defaultWorkflowLogger(ctx, option)

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

	logger.Info("Checking for the existence of Workspace instance")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		return fmt.Errorf("workspace instance not found: %w", err)
	}

	logger.Info("Stopping Workspace instance")
	var opName string
	if err := workflow.ExecuteActivity(ctx, wa.Stop, option).Get(ctx, &opName); err != nil {
		return fmt.Errorf("failed to stop Workspace instance: %w", err)
	}

	logger.Info("Waiting for Workspace instance stopped")
	if err := workflow.ExecuteActivity(ctx, wa.OperationCompleted, opName).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to watch operation to stop Workspace instance: %w", err)
	}

	logger.Info("Workspace instance stopped successfully!")
	return nil
}
