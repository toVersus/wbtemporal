package jupyterhubapi

import (
	"context"
	"fmt"
)

const (
	ExecutorNameJupyterHub = "jupyterhub"
	ExecutorNameFakeClient = "fakeclient"
)

var (
	ErrNotFoundExecutor = fmt.Errorf("executor not found")
)

type Option struct {
	// Server indicates the name of user server
	Server string
	// User indicates the owner of user server
	User string
}

type Status struct {
	Name   string
	URL    string
	Status string
}

// NotebookService is an interface for interacting with Google Cloud Notebooks API
type NotebookService interface {
	DescribeUserServer(ctx context.Context, option *Option) (*Status, error)
	CreateUserServer(ctx context.Context, option *Option) error
	DeleteUserServer(ctx context.Context, option *Option) error
	IsUserServerReady(ctx context.Context, option *Option) (bool, error)
	IsUserServerDeleted(ctx context.Context, option *Option) (bool, error)
}

type Executor interface {
	NotebookService
}
