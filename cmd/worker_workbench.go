package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toVersus/wbtemporal/pkg/executor/googleapi"
)

var (
	workerWorkbenchCmd = &cobra.Command{
		Use: "workbench",
	}
)

func NewGoogleAPIExecutor(ctx context.Context, opts ExecutorOpts) (googleapi.Executor, error) {
	if opts.Name == googleapi.ExecutorNameGoogleAPI {
		return googleapi.NewWorkbench(ctx)
	}
	// } else if opts.Name == executor.ExecutorNameFakeClient {
	// 	return fakeclient.NewFakeClientExecutor(), nil
	// }
	return nil, fmt.Errorf("executor %s not supported: %w", opts.Name, ErrNotFoundExecutor)
}
