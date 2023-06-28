package cmd

import (
	"context"

	"github.com/toVersus/wbtemporal/pkg/logger"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

type workflowWatcher struct {
	c  client.Client
	id string
}

func (w *workflowWatcher) run(ctx context.Context) {
	logger := logger.NewDefaultLogger(logLevel)

	workflowId := w.id
	var runId string
	for {
		workflow, err := w.c.DescribeWorkflowExecution(ctx, workflowId, "")
		if err != nil {
			logger.Fatal("Could not describe latest workflow", "Error", err)
		}
		info := workflow.GetWorkflowExecutionInfo()
		if info.Status == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
			runId = info.Execution.RunId
			logger.Info("Start watching event for running workflow",
				"runid", runId,
			)
			break
		}
	}
	iter := w.c.GetWorkflowHistory(ctx, workflowId, runId, true, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var workflowName, activityName string
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			switch err.(type) {
			case *serviceerror.NotFound:
				continue
			default:
				logger.Fatal("Could not retrieve latest workflow execution history", "Error", err)
			}
		}
		switch event.EventType {
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
			workflow := event.GetWorkflowExecutionStartedEventAttributes()
			workflowName = workflow.WorkflowType.Name
			logger.Info("Workflow started",
				"eventtime", event.EventTime,
				"workflow", workflowName,
				"workflowId", workflowId,
			)
		case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			activity := event.GetActivityTaskScheduledEventAttributes()
			activityName = activity.ActivityType.Name
			logger.Info("Activity scheduled",
				"eventtime", event.EventTime,
				"activity", activityName,
				"workflowId", workflowId,
			)
		case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
			logger.Info("Activity started",
				"eventtime", event.EventTime,
				"activity", activityName,
				"workflowId", workflowId,
			)
		case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
			logger.Info("Activity completed successfully",
				"eventtime", event.EventTime,
				"activity", activityName,
				"workflowId", workflowId,
			)
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
			logger.Info("Workflow completed successfully",
				"eventtime", event.EventTime,
				"workflow", workflowName,
				"workflowId", workflowId,
			)
		default:
			continue
		}
	}
}
