package workflow

import (
	"github.com/toVersus/wbtemporal/pkg/executor"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

func defaultGoogleAPIWorkflowLogger(ctx workflow.Context, workspace *executor.WorkspaceOption) log.Logger {
	return log.With(workflow.GetLogger(ctx),
		"ProjectID", workspace.GoogleAPIOption.ProjectId,
		"ClusterName", workspace.Name,
		"ClusterLocation", workspace.GoogleAPIOption.Location,
	)
}
