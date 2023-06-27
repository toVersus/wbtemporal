package googleapi

import (
	"context"
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	notebooks "cloud.google.com/go/notebooks/apiv1"
	"cloud.google.com/go/notebooks/apiv1/notebookspb"
	"github.com/toVersus/wbtemporal/pkg/executor"
)

var (
	_ executor.NotebookService             = &workbench{}
	_ executor.LongRunningOperationService = &workbench{}
)

type workbench struct {
	notebookClient *notebooks.NotebookClient
}

func NewWorkbench(ctx context.Context) (executor.Executor, error) {
	notebookClient, err := notebooks.NewNotebookClient(ctx)
	if err != nil {
		return &workbench{}, fmt.Errorf("failed to initialize compute service: %s", err)
	}

	return &workbench{
		notebookClient: notebookClient,
	}, nil
}

func (w *workbench) CreateNotebookInstance(ctx context.Context, option *executor.WorkspaceOption) (string, error) {
	req := &notebookspb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s/locations/%s", option.ProjectId, option.Zone),
		InstanceId: option.Name,
		Instance: &notebookspb.Instance{
			Environment: &notebookspb.Instance_VmImage{
				&notebookspb.VmImage{
					Project: "deeplearning-platform-release",
					Image: &notebookspb.VmImage_ImageFamily{
						ImageFamily: "common-cpu-notebooks",
					},
				},
			},
			BootDiskType: notebookspb.Instance_PD_BALANCED,
			// 最小構成で 50 GB は必要
			BootDiskSizeGb: 50,
			DataDiskType:   notebookspb.Instance_PD_BALANCED,
			DataDiskSizeGb: 20,
			Network:        fmt.Sprintf("projects/%s/global/networks/%s", option.ProjectId, option.Network),
			Subnet:         fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", option.ProjectId, option.Location, option.Subnet),
			InstanceOwners: []string{option.Email},
			MachineType:    option.MachineType,
		},
	}
	op, err := w.notebookClient.CreateInstance(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create user managed notebook instance: %w", err)
	}
	return op.Name(), nil
}

func (w *workbench) DescribeNotebookInstance(ctx context.Context, projectID, zone, name string) (*executor.WorkspaceStatus, error) {
	req := &notebookspb.GetInstanceRequest{
		Name: notebookInstanceFullname(projectID, zone, name),
	}
	wb, err := w.notebookClient.GetInstance(ctx, req)
	if err != nil {
		return nil, err
	}
	return &executor.WorkspaceStatus{
		Name:   wb.Name,
		URL:    wb.ProxyUri,
		Status: wb.State.String(),
	}, nil
}

func (w *workbench) StartNotebookInstance(ctx context.Context, projectID, zone, name string) (string, error) {
	op, err := w.notebookClient.StartInstance(ctx, &notebookspb.StartInstanceRequest{
		Name: notebookInstanceFullname(projectID, zone, name),
	})
	if err != nil {
		return "", err
	}
	return op.Name(), nil
}

func (w *workbench) StopNotebookInstance(ctx context.Context, projectID, zone, name string) (string, error) {
	op, err := w.notebookClient.StopInstance(ctx, &notebookspb.StopInstanceRequest{
		Name: notebookInstanceFullname(projectID, zone, name),
	})
	if err != nil {
		return "", err
	}
	return op.Name(), nil
}

func (w *workbench) DeleteNotebookInstance(ctx context.Context, projectID, zone, name string) (string, error) {
	op, err := w.notebookClient.DeleteInstance(ctx, &notebookspb.DeleteInstanceRequest{
		Name: notebookInstanceFullname(projectID, zone, name),
	})
	if err != nil {
		return "", err
	}
	return op.Name(), nil
}

func (w workbench) HasOperationDone(ctx context.Context, name string) (bool, error) {
	op, err := w.notebookClient.GetOperation(ctx, &longrunningpb.GetOperationRequest{
		Name: name,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get notebook operation %q: %w", name, err)
	}
	if op.GetError() != nil {
		return false, fmt.Errorf("notebook operation %q aborted: %s", name, op.GetError().GetMessage())
	}
	if !op.GetDone() {
		return false, nil
	}
	return true, nil
}

func notebookInstanceFullname(projectID, zone, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectID, zone, name)
}
