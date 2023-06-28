package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/executor/jupyterhubapi"
)

var (
	workerJupyterHubCmd = &cobra.Command{
		Use: "jupyterhub",
	}
)

func NewJupyterHubExecutor(ctx context.Context, opts ExecutorOpts) (jupyterhubapi.Executor, error) {
	if opts.Name == jupyterhubapi.ExecutorNameJupyterHub {
		if len(jupyterHubBaseURL) == 0 {
			return nil, fmt.Errorf("jupyterhub base url is required")
		}
		if len(jupyterHubAPIToken) == 0 {
			return nil, fmt.Errorf("jupyterhub api token is required")
		}
		return jupyterhubapi.NewExecutor(ctx, jupyterHubBaseURL, jupyterHubAPIToken)
	}
	// } else if opts.Name == executor.ExecutorNameFakeClient {
	// 	return fakeclient.NewFakeClientExecutor(), nil
	// }
	return nil, fmt.Errorf("executor %s not supported: %w", opts.Name, ErrNotFoundExecutor)
}
