package workflow

import (
	"fmt"
	"time"

	"github.com/toVersus/wbtemporal/pkg/activity"
	"github.com/toVersus/wbtemporal/pkg/executor/jupyterhubapi"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	CreateJupyterHubTaskQueue = "CREATE_JUPYTERHUB_TASK_QUEUE"
	DeleteJupyterHubTaskQueue = "DELETE_JUPYTERHUB_TASK_QUEUE"
)

func CreateUserServer(ctx workflow.Context, option *jupyterhubapi.Option) (*jupyterhubapi.Status, error) {
	var wa *activity.JupyterHubActivity

	logger := defaultJupyterHubWorkflowLogger(ctx, option)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// アクティビティの実行時間のタイムアウト値
		StartToCloseTimeout: 1 * time.Minute,
		// アクティビティを 5 秒間隔で 72 回の合計 6 分間リトライする
		// JupyterHub の user server の作成を待つ時のリトライ戦略をベースに設定
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        5 * time.Second,
			MaximumInterval:        5 * time.Second,
			MaximumAttempts:        72,
			NonRetryableErrorTypes: []string{activity.ErrLongRunningOperationFailed},
		},
	})

	logger.Info("Checking for the existence of user server")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		return nil, fmt.Errorf("failed to check for the existence of user server: %w", err)
	}

	if exist {
		logger.Info("User server already exists")
	} else {
		logger.Info("Creating new user server")
		if err := workflow.ExecuteActivity(ctx, wa.Create, option).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("failed to create user server: %w", err)
		}

		logger.Info("Waiting for user server to become ready")
		if err := workflow.ExecuteActivity(ctx, wa.WaitReady, option).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("failed to watch operation for creation of Notebook instance: %w", err)
		}
	}

	var status jupyterhubapi.Status
	logger.Info("Getting access URL for user server")
	if err := workflow.ExecuteActivity(ctx, wa.GetAccessURL, option).Get(ctx, &status); err != nil {
		return nil, fmt.Errorf("failed to watch operation to create user server: %w", err)
	}

	logger.Info("User server created successfully!")
	return &status, nil
}

func DeleteUserServer(ctx workflow.Context, option *jupyterhubapi.Option) error {
	var wa *activity.JupyterHubActivity

	logger := defaultJupyterHubWorkflowLogger(ctx, option)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// アクティビティの実行時間のタイムアウト値
		StartToCloseTimeout: 1 * time.Minute,
		// アクティビティを 5 秒間隔で 36 回の合計 3 分間リトライする
		// JupyterHub の user server の削除を待つ時のリトライ戦略をベースに設定
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        5 * time.Second,
			MaximumInterval:        5 * time.Second,
			MaximumAttempts:        36,
			NonRetryableErrorTypes: []string{activity.ErrLongRunningOperationFailed},
		},
	})

	logger.Info("Checking for the existence of user server")
	var exist bool
	if err := workflow.ExecuteActivity(ctx, wa.Exist, option).Get(ctx, &exist); err != nil {
		logger.Info("User server already not exists")
		return nil
	}
	if !exist {
		logger.Info("User server already deleted")
	}

	logger.Info("Deleting user server")
	if err := workflow.ExecuteActivity(ctx, wa.Delete, option).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to delete user server: %w", err)
	}

	logger.Info("Waiting for user server deleted")
	if err := workflow.ExecuteActivity(ctx, wa.WaitDeleted, option).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to watch operation to delete user server: %w", err)
	}

	logger.Info("User server deleted successfully!")
	return nil
}
