package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)

func TestWaitExecutor(t *testing.T) {
	ex := executor.WaitExecutor{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := ex.Execute(ctx, model.Job{
		Type:    model.JobWait,
		Payload: map[string]any{"duration_seconds": 1},
	})
	if result.Status != model.JobSuccess {
		t.Fatalf("expected success, got %s: %s", result.Status, result.Error)
	}
}

func TestWriteAndFileExistsExecutors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	writeEx := executor.WriteFileExecutor{}
	result := writeEx.Execute(context.Background(), model.Job{
		Type:    model.JobWriteFile,
		Payload: map[string]any{"path": path, "content": "hello"},
	})
	if result.Status != model.JobSuccess {
		t.Fatalf("write failed: %s", result.Error)
	}

	existsEx := executor.FileExistsExecutor{}
	result = existsEx.Execute(context.Background(), model.Job{
		Type:    model.JobFileExists,
		Payload: map[string]any{"path": path},
	})
	if result.Status != model.JobSuccess {
		t.Fatalf("file_exists failed: %s", result.Error)
	}
}

func TestChecksumExecutor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}

	ex := executor.ChecksumExecutor{}
	result := ex.Execute(context.Background(), model.Job{
		Type:    model.JobChecksum,
		Payload: map[string]any{"path": path},
	})
	if result.Status != model.JobSuccess {
		t.Fatalf("checksum failed: %s", result.Error)
	}
	if result.Result["checksum"] == "" {
		t.Fatal("expected checksum in result")
	}
}

func TestCopyFileExecutor(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.txt")
	dest := filepath.Join(dir, "dest.txt")
	if err := os.WriteFile(source, []byte("copy me"), 0o644); err != nil {
		t.Fatal(err)
	}

	ex := executor.CopyFileExecutor{}
	result := ex.Execute(context.Background(), model.Job{
		Type: model.JobCopyFile,
		Payload: map[string]any{
			"source":      source,
			"destination": dest,
		},
	})
	if result.Status != model.JobSuccess {
		t.Fatalf("copy failed: %s", result.Error)
	}
	data, err := os.ReadFile(dest)
	if err != nil || string(data) != "copy me" {
		t.Fatalf("unexpected copied content: %q err=%v", data, err)
	}
}

func TestDefaultRegistry(t *testing.T) {
	reg := executor.DefaultRegistry()
	types := []model.JobType{
		model.JobHTTPCheck,
		model.JobTCPCheck,
		model.JobFileExists,
		model.JobChecksum,
		model.JobCopyFile,
		model.JobWriteFile,
		model.JobWait,
	}
	for _, jobType := range types {
		if _, ok := reg.Get(jobType); !ok {
			t.Fatalf("missing executor for %s", jobType)
		}
	}
}
