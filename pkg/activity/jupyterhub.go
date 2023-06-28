package activity

import (
	"context"
	"fmt"
	"strings"

	"github.com/toVersus/wbtemporal/pkg/executor/jupyterhubapi"
	"go.temporal.io/sdk/temporal"
)

const (
	ErrOperationFailed = "ErrorOperationFailed"
)

type JupyterHubActivity struct {
	Executor jupyterhubapi.Executor
}

func (a *JupyterHubActivity) Exist(ctx context.Context, option *jupyterhubapi.Option) (bool, error) {
	_, err := a.Executor.DescribeUserServer(ctx, option)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *JupyterHubActivity) Create(ctx context.Context, option *jupyterhubapi.Option) error {
	err := a.Executor.CreateUserServer(ctx, option)
	if err != nil {
		return err
	}
	return nil
}

func (a *JupyterHubActivity) Delete(ctx context.Context, option *jupyterhubapi.Option) error {
	err := a.Executor.DeleteUserServer(ctx, option)
	if err != nil {
		return err
	}
	return nil
}

func (a *JupyterHubActivity) GetAccessURL(ctx context.Context, option *jupyterhubapi.Option) (*jupyterhubapi.Status, error) {
	result, err := a.Executor.DescribeUserServer(ctx, option)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *JupyterHubActivity) WaitReady(ctx context.Context, option *jupyterhubapi.Option) error {
	ready, err := a.Executor.IsUserServerReady(ctx, option)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("non-retryable error found in waiting to become ready", ErrOperationFailed, err)
	}
	if !ready {
		return fmt.Errorf("instance is not ready yet")
	}
	return nil
}

func (a *JupyterHubActivity) WaitDeleted(ctx context.Context, option *jupyterhubapi.Option) error {
	ready, err := a.Executor.IsUserServerDeleted(ctx, option)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("non-retryable error found in waiting to be deleted", ErrOperationFailed, err)
	}
	if !ready {
		return fmt.Errorf("instance is not deleted yet")
	}
	return nil
}
