package executor

import (
	"context"
	"os"

	"taskbridge/internal/model"
)

type FileExistsExecutor struct{}

func (FileExistsExecutor) Type() model.JobType { return model.JobFileExists }

func (FileExistsExecutor) Execute(ctx context.Context, job model.Job) Result {
	path, _ := job.Payload["path"].(string)
	if path == "" {
		return failResult("path is required")
	}

	select {
	case <-ctx.Done():
		return failResult(ctx.Err().Error())
	default:
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return failResult("file does not exist")
	}
	if err != nil {
		return failResult(err.Error())
	}

	return successResult([]string{"file exists: " + path}, map[string]any{
		"path":   path,
		"exists": true,
	})
}
