package workflow

import (
	"github.com/toVersus/wbtemporal/pkg/executor/googleapi"
	"github.com/toVersus/wbtemporal/pkg/executor/jupyterhubapi"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

func defaultGoogleAPIWorkflowLogger(ctx workflow.Context, option *googleapi.Option) log.Logger {
	return log.With(workflow.GetLogger(ctx),
		"ProjectID", option.ProjectId,
		"ClusterName", option.Name,
		"ClusterLocation", option.Location,
	)
}

func defaultJupyterHubWorkflowLogger(ctx workflow.Context, option *jupyterhubapi.Option) log.Logger {
	return log.With(workflow.GetLogger(ctx),
		"User", option.User,
		"Server", option.Server,
	)
}
