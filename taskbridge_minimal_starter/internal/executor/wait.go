package executor

import (
	"context"
	"time"

	"taskbridge/internal/model"
)

type WaitExecutor struct{}

func (WaitExecutor) Type() model.JobType { return model.JobWait }

func (WaitExecutor) Execute(ctx context.Context, job model.Job) Result {
	seconds := 1
	if v, ok := job.Payload["duration_seconds"]; ok {
		switch n := v.(type) {
		case float64:
			seconds = int(n)
		case int:
			seconds = n
		}
	}
	if seconds < 0 {
		return failResult("duration_seconds must be non-negative")
	}

	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return failResult(ctx.Err().Error())
	case <-timer.C:
	}

	return successResult([]string{"wait completed"}, map[string]any{
		"duration_seconds": seconds,
	})
}
