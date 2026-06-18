package executor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"taskbridge/internal/model"
)

type ChecksumExecutor struct{}
func (ChecksumExecutor) Type() model.JobType { return model.JobChecksum }

func (ChecksumExecutor) Execute(ctx context.Context, job model.Job) Result {
	path, _ := job.Payload["path"].(string)
	if path == "" {
		return failResult("path is required")
	}
	f, err := os.Open(path)
	if err != nil {
		return failResult(err.Error())
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return failResult(err.Error())
	}

	sum := hex.EncodeToString(h.Sum(nil))
	return successResult([]string{"checksum computed for " + path}, map[string]any{
		"path":      path,
		"checksum":  sum,
		"algorithm": "sha256",
	})
}
