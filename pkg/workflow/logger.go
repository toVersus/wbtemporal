package workflow

import (
	"github.com/toVersus/wbtemporal/pkg/executor"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

func defaultWorkflowLogger(ctx workflow.Context, workspace *executor.WorkspaceOption) log.Logger {
	return log.With(workflow.GetLogger(ctx),
		"ProjectID", workspace.ProjectId,
		"ClusterName", workspace.Name,
		"ClusterLocation", workspace.Location,
	)
}
