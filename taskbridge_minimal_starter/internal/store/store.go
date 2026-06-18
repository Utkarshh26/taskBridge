package store

import "taskbridge/internal/model"

type Store interface {
	CreateJob(job model.Job) (model.Job, error)
	ListJobs() ([]model.Job, error)
	GetJob(jobID string) (model.Job, bool, error)
	UpdateJob(job model.Job) error

	RegisterAgent(agent model.Agent) (model.Agent, error)
	Heartbeat(agentID string) error
	ListAgents() ([]model.Agent, error)
	GetAgent(agentID string) (model.Agent, bool, error)

	AssignNextJob(agentID string, capabilities []model.JobType) (model.Job, bool, error)
	CompleteJob(jobID string, status model.JobStatus, logs []string, result map[string]any, errMsg string) error
}
