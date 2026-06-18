package executor

import (
	"context"
	"os"

	"taskbridge/internal/model"
)

type WriteFileExecutor struct{}

func (WriteFileExecutor) Type() model.JobType { return model.JobWriteFile }

func (WriteFileExecutor) Execute(ctx context.Context, job model.Job) Result {
	path, _ := job.Payload["path"].(string)
	content, _ := job.Payload["content"].(string)
	if path == "" {
		return failResult("path is required")
	}

	select {
	case <-ctx.Done():
		return failResult(ctx.Err().Error())
	default:
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return failResult(err.Error())
	}

	return successResult([]string{"wrote file: " + path}, map[string]any{
		"path":  path,
		"bytes": len(content),
	})
}
