package activity

import (
	"context"
	"fmt"
	"strings"

	"github.com/toVersus/wbtemporal/pkg/executor"
	"go.temporal.io/sdk/temporal"
)

const (
	ErrLongRunningOperationFailed = "ErrorLongRunningOperationFailed"
)

type WorkspaceActvity struct {
	Executor executor.Executor
}

func (w *WorkspaceActvity) Exist(ctx context.Context, option *executor.WorkspaceOption) (bool, error) {
	_, err := w.Executor.DescribeNotebookInstance(ctx, option)
	if err != nil {
		if strings.Contains(err.Error(), "was not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (w *WorkspaceActvity) GetWorkspaceURL(ctx context.Context, option *executor.WorkspaceOption) (*executor.WorkspaceStatus, error) {
	result, err := w.Executor.DescribeNotebookInstance(ctx, option)
	if err != nil {
		return nil, err
	}
	// Workspace を作成する Operation はあくまで Workspace Instance を作成するまでしか待たないので、
	// Workspace Instance が Active になるまで待つために、接続先の URL が取得できるまで待つ
	if len(result.URL) == 0 {
		return nil, fmt.Errorf("workspace instance is not active yet")
	}

	return result, nil
}

func (w *WorkspaceActvity) Create(ctx context.Context, option *executor.WorkspaceOption) (string, error) {
	opName, err := w.Executor.CreateNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (w *WorkspaceActvity) Delete(ctx context.Context, option *executor.WorkspaceOption) (string, error) {
	opName, err := w.Executor.DeleteNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (w *WorkspaceActvity) Start(ctx context.Context, option *executor.WorkspaceOption) (string, error) {
	opName, err := w.Executor.StartNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (w *WorkspaceActvity) Stop(ctx context.Context, option *executor.WorkspaceOption) (string, error) {
	opName, err := w.Executor.StopNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (w *WorkspaceActvity) OperationCompleted(ctx context.Context, opName string) error {
	done, err := w.Executor.HasOperationDone(ctx, opName)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("non-retryable error found in watch operation", ErrLongRunningOperationFailed, err)
	}
	if !done {
		return fmt.Errorf("operation is not done yet")
	}
	return nil
}
