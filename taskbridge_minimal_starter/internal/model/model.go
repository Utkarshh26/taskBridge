package model

import "time"

type JobStatus string

const (
	JobPending  JobStatus = "PENDING"
	JobRunning  JobStatus = "RUNNING"
	JobRetrying JobStatus = "RETRYING"
	JobSuccess  JobStatus = "SUCCESS"
	JobFailed   JobStatus = "FAILED"
)

type JobType string

const (
	JobHTTPCheck  JobType = "http_check"
	JobTCPCheck   JobType = "tcp_check"
	JobFileExists JobType = "file_exists"
	JobChecksum   JobType = "checksum"
	JobCopyFile   JobType = "copy_file"
	JobWriteFile  JobType = "write_file"
	JobWait       JobType = "wait"
)

type Job struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            JobType        `json:"type"`
	Payload         map[string]any `json:"payload"`
	Status          JobStatus      `json:"status"`
	AssignedAgentID string         `json:"assigned_agent_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	FinishedAt      *time.Time     `json:"finished_at,omitempty"`
	AttemptCount    int            `json:"attempt_count"`
	MaxRetries      int            `json:"max_retries"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	Logs            []string       `json:"logs,omitempty"`
	Error           string         `json:"error,omitempty"`
	Result          map[string]any `json:"result,omitempty"`
}

type Agent struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	Capabilities []JobType `json:"capabilities"`
	LastSeen     time.Time `json:"last_seen"`
	Status       string    `json:"status"`
}

const AgentStatusOnline = "online"
