package executor

import (
	"context"
	"io"
	"os"

	"taskbridge/internal/model"
)

type CopyFileExecutor struct{}

func (CopyFileExecutor) Type() model.JobType { return model.JobCopyFile }

func (CopyFileExecutor) Execute(ctx context.Context, job model.Job) Result {
	source, _ := job.Payload["source"].(string)
	dest, _ := job.Payload["destination"].(string)
	if source == "" || dest == "" {
		return failResult("source and destination are required")
	}

	select {
	case <-ctx.Done():
		return failResult(ctx.Err().Error())
	default:
	}

	in, err := os.Open(source)
	if err != nil {
		return failResult(err.Error())
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return failResult(err.Error())
	}

	written, err := io.Copy(out, in)
	closeErr := out.Close()
	if err != nil {
		return failResult(err.Error())
	}
	if closeErr != nil {
		return failResult(closeErr.Error())
	}

	return successResult([]string{"copied file"}, map[string]any{
		"source":      source,
		"destination": dest,
		"bytes":       written,
	})
}
