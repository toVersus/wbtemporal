package activity

import (
	"context"
	"fmt"
	"strings"

	"github.com/toVersus/wbtemporal/pkg/executor/googleapi"
	"go.temporal.io/sdk/temporal"
)

const (
	ErrLongRunningOperationFailed = "ErrorLongRunningOperationFailed"
)

type WorkbenchActivity struct {
	Executor googleapi.Executor
}

func (a *WorkbenchActivity) Exist(ctx context.Context, option *googleapi.Option) (bool, error) {
	_, err := a.Executor.DescribeNotebookInstance(ctx, option)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *WorkbenchActivity) GetWorkspaceURL(ctx context.Context, option *googleapi.Option) (*googleapi.Status, error) {
	result, err := a.Executor.DescribeNotebookInstance(ctx, option)
	if err != nil {
		return nil, err
	}
	// Vertex AI Workbench Instance を作成する Operation はあくまで Workbench Instance を作成するまでしか待たないので、
	// Workbench Instance が Active になるまで待つために、接続先の URL が取得できるまで待つ
	if len(result.URL) == 0 {
		return nil, fmt.Errorf("workbench instance is not active yet")
	}

	return result, nil
}

func (a *WorkbenchActivity) Create(ctx context.Context, option *googleapi.Option) (string, error) {
	opName, err := a.Executor.CreateNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (a *WorkbenchActivity) Delete(ctx context.Context, option *googleapi.Option) (string, error) {
	opName, err := a.Executor.DeleteNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (a *WorkbenchActivity) Start(ctx context.Context, option *googleapi.Option) (string, error) {
	opName, err := a.Executor.StartNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (a *WorkbenchActivity) Stop(ctx context.Context, option *googleapi.Option) (string, error) {
	opName, err := a.Executor.StopNotebookInstance(ctx, option)
	if err != nil {
		return "", err
	}
	return opName, nil
}

func (a *WorkbenchActivity) OperationCompleted(ctx context.Context, opName string) error {
	done, err := a.Executor.HasOperationDone(ctx, opName)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("non-retryable error found in watch operation", ErrLongRunningOperationFailed, err)
	}
	if !done {
		return fmt.Errorf("operation is not done yet")
	}
	return nil
}
