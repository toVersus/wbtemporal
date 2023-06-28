package googleapi

import (
	"context"
	"fmt"
)

const (
	ExecutorNameGoogleAPI  = "googleapi"
	ExecutorNameFakeClient = "fakeclient"
)

var (
	ErrNotFoundExecutor = fmt.Errorf("executor not found")
)

type Option struct {
	// Name indicates the workspace name
	Name string
	// Email indicates the workspace owner email
	Email string
	// Region indicates the network region
	Zone string
	// Location indicates the workspace location or zone
	Location string
	// ProjectId indicates the GCP project ID.
	// There is a way to automatically discover the project ID using google SDK from credentials,
	// but we don't support this feature at this momenet. Users must explicitly set GCP project ID.
	// https://pkg.go.dev/golang.org/x/oauth2/google#FindDefaultCredentials
	ProjectId string
	// MachineType indicates the workspace machine type
	MachineType string
	// Network indicates the VPC network that workspace instances are deployed to
	Network string
	// Subnet indicates the subnet that workspace instances are deployed to
	Subnet string
}

type Status struct {
	Name   string
	URL    string
	Status string
}

// NotebookService is an interface for interacting with Google Cloud Notebooks API
type NotebookService interface {
	CreateNotebookInstance(ctx context.Context, option *Option) (string, error)
	DescribeNotebookInstance(ctx context.Context, option *Option) (*Status, error)
	StartNotebookInstance(ctx context.Context, option *Option) (string, error)
	StopNotebookInstance(ctx context.Context, option *Option) (string, error)
	DeleteNotebookInstance(ctx context.Context, option *Option) (string, error)
}

type LongRunningOperationService interface {
	HasOperationDone(ctx context.Context, opName string) (bool, error)
}

type Executor interface {
	NotebookService
	LongRunningOperationService
}
