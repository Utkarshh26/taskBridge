package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"taskbridge/internal/model"
)

type HTTPCheckExecutor struct{}

func (HTTPCheckExecutor) Type() model.JobType { return model.JobHTTPCheck }

func (HTTPCheckExecutor) Execute(ctx context.Context, job model.Job) Result {
	url, _ := job.Payload["url"].(string)
	if url == "" {
		return failResult("url is required")
	}

	expected := 200
	if v, ok := job.Payload["expected_status"]; ok {
		switch n := v.(type) {
		case float64:
			expected = int(n)
		case int:
			expected = n
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return failResult(err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return failResult(err.Error())
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != expected {
		return failResult(fmt.Sprintf("expected status %d, got %d", expected, resp.StatusCode))
	}

	return successResult([]string{fmt.Sprintf("GET %s returned %d", url, resp.StatusCode)}, map[string]any{
		"status_code": resp.StatusCode,
	})
}
