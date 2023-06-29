package activity

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/temporal"

	"github.com/toVersus/wbtemporal/pkg/client/jupyterhub"
	"github.com/toVersus/wbtemporal/pkg/executor/jupyterhubapi"
)

const (
	ErrOperationFailed = "ErrorOperationFailed"
)

type JupyterHubActivity struct {
	Executor jupyterhubapi.Executor
}

func (a *JupyterHubActivity) GetOrCreateUser(ctx context.Context, option *jupyterhubapi.Option) (*jupyterhub.User, error) {
	user, err := a.Executor.GetUser(ctx, option)
	if err == jupyterhubapi.ErrUserNotFound {
		user, err = a.Executor.CreateUser(ctx, option)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return user, nil
}

func (a *JupyterHubActivity) ExistUserServer(ctx context.Context, option *jupyterhubapi.Option) (bool, error) {
	server, err := a.Executor.GetUserServer(ctx, option)
	if err == jupyterhubapi.ErrServerNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return server.Status == jupyterhubapi.UserServerStatusReady, nil
}

func (a *JupyterHubActivity) CreateUserServer(ctx context.Context, option *jupyterhubapi.Option) error {
	err := a.Executor.CreateUserServer(ctx, option)
	if err != nil {
		return err
	}
	return nil
}

func (a *JupyterHubActivity) DeleteUserServer(ctx context.Context, option *jupyterhubapi.Option) error {
	err := a.Executor.DeleteUserServer(ctx, option)
	if err != nil {
		return err
	}
	return nil
}

func (a *JupyterHubActivity) GetUserServer(ctx context.Context, option *jupyterhubapi.Option) (*jupyterhubapi.Status, error) {
	server, err := a.Executor.GetUserServer(ctx, option)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (a *JupyterHubActivity) WaitUserServerReady(ctx context.Context, option *jupyterhubapi.Option) error {
	ready, err := a.Executor.IsUserServerReady(ctx, option)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("non-retryable error found in waiting to become ready", ErrOperationFailed, err)
	}
	if !ready {
		return fmt.Errorf("instance is not ready yet")
	}
	return nil
}

func (a *JupyterHubActivity) WaitUserServerDeleted(ctx context.Context, option *jupyterhubapi.Option) error {
	ready, err := a.Executor.IsUserServerDeleted(ctx, option)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("non-retryable error found in waiting to be deleted", ErrOperationFailed, err)
	}
	if !ready {
		return fmt.Errorf("instance is not deleted yet")
	}
	return nil
}
