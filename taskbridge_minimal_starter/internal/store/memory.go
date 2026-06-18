package store

import (
	"time"

	"taskbridge/internal/model"
	"taskbridge/internal/scheduler"
)

type MemoryStore struct {
	jobs   *JobStore
	agents *AgentStore
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs:   NewJobStore(),
		agents: NewAgentStore(),
	}
}

func (s *MemoryStore) CreateJob(job model.Job) (model.Job, error) {
	return s.jobs.Create(job)
}

func (s *MemoryStore) ListJobs() ([]model.Job, error) {
	return s.jobs.List(), nil
}

func (s *MemoryStore) GetJob(jobID string) (model.Job, bool, error) {
	job, ok := s.jobs.Get(jobID)
	return job, ok, nil
}

func (s *MemoryStore) UpdateJob(job model.Job) error {
	return s.jobs.Update(job)
}

func (s *MemoryStore) RegisterAgent(agent model.Agent) (model.Agent, error) {
	return s.agents.Register(agent)
}

func (s *MemoryStore) Heartbeat(agentID string) error {
	agent, ok := s.agents.Get(agentID)
	if !ok {
		return ErrAgentNotFound
	}
	agent.LastSeen = time.Now().UTC()
	agent.Status = model.AgentStatusOnline
	return s.agents.Update(agent)
}

func (s *MemoryStore) ListAgents() ([]model.Agent, error) {
	return s.agents.List(), nil
}

func (s *MemoryStore) GetAgent(agentID string) (model.Agent, bool, error) {
	agent, ok := s.agents.Get(agentID)
	return agent, ok, nil
}

func (s *MemoryStore) AssignNextJob(agentID string, capabilities []model.JobType) (model.Job, bool, error) {
	job, ok := s.jobs.AssignNext(agentID, capabilities)
	if !ok {
		return model.Job{}, false, nil
	}
	now := time.Now().UTC()
	job.StartedAt = &now
	if err := s.jobs.Update(job); err != nil {
		return model.Job{}, false, err
	}
	return job, true, nil
}

func (s *MemoryStore) CompleteJob(jobID string, status model.JobStatus, logs []string, result map[string]any, errMsg string) error {
	job, ok := s.jobs.Get(jobID)
	if !ok {
		return ErrJobNotFound
	}

	now := time.Now().UTC()
	job.Logs = logs
	job.Result = result
	job.Error = errMsg
	job.FinishedAt = &now
	job.AssignedAgentID = ""

	finalStatus := scheduler.ResolveCompletionStatus(job, status)
	job.Status = finalStatus

	return s.jobs.Complete(jobID, job)
}
