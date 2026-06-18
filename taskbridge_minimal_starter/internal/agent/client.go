package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"taskbridge/internal/executor"
	"taskbridge/internal/model"
)
type Client struct {
	baseURL    string
	agentID    string
	httpClient *http.Client
	registry   *executor.Registry
}

type Config struct {
	ServerURL    string
	AgentID      string
	Capabilities []model.JobType
	Hostname     string
}

func NewClient(cfg Config, registry *executor.Registry) *Client {
	if registry == nil {
		registry = executor.DefaultRegistry()
	}
	return &Client{
		baseURL:  strings.TrimRight(cfg.ServerURL, "/"),
		agentID:  cfg.AgentID,
		registry: registry,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Register(ctx context.Context, hostname string, capabilities []model.JobType) (model.Agent, error) {
	req := model.RegisterAgentRequest{
		Hostname:     hostname,
		Capabilities: capabilities,
	}
	var resp model.RegisterAgentResponse
	if err := c.post(ctx, "/agents/register", req, &resp); err != nil {
		return model.Agent{}, err
	}
	c.agentID = resp.Agent.ID
	return resp.Agent, nil
}

func (c *Client) Heartbeat(ctx context.Context) error {
	return c.post(ctx, fmt.Sprintf("/agents/%s/heartbeat", c.agentID), map[string]any{}, nil)
}

func (c *Client) PollNextJob(ctx context.Context) (*model.Job, error) {
	var resp model.NextJobResponse
	if err := c.post(ctx, fmt.Sprintf("/agents/%s/next-job", c.agentID), map[string]any{}, &resp); err != nil {
		return nil, err
	}
	if !resp.Found {
		return nil, nil
	}
	return resp.Job, nil
}

func (c *Client) SubmitResult(ctx context.Context, jobID string, result executor.Result) error {
	req := model.JobResultRequest{
		Status: result.Status,
		Logs:   result.Logs,
		Result: result.Result,
		Error:  result.Error,
	}
	return c.post(ctx, fmt.Sprintf("/jobs/%s/result", jobID), req, nil)
}

func (c *Client) Run(ctx context.Context, pollInterval, heartbeatInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-heartbeatTicker.C:
			if err := c.Heartbeat(ctx); err != nil {
				return err
			}
		case <-ticker.C:
			job, err := c.PollNextJob(ctx)
			if err != nil {
				return err
			}
			if job == nil {
				continue
			}
			if err := c.executeAndReport(ctx, *job); err != nil {
				return err
			}
		}
	}
}

func (c *Client) executeAndReport(ctx context.Context, job model.Job) error {
	ex, ok := c.registry.Get(job.Type)
	if !ok {
		result := executor.Result{
			Status: model.JobFailed,
			Logs:   []string{"unsupported job type"},
			Error:  "unsupported job type: " + string(job.Type),
		}
		return c.SubmitResult(ctx, job.ID, result)
	}

	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := ex.Execute(runCtx, job)
	return c.SubmitResult(ctx, job.ID, result)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var errResp model.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}
	if out != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func Hostname() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "unknown"
	}
	return name
}

func ParseCapabilities(raw string) []model.JobType {
	parts := strings.Split(raw, ",")
	out := make([]model.JobType, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, model.JobType(p))
	}
	return out
}
