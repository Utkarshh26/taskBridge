package store

import (
	"sort"
	"sync"

	"taskbridge/internal/model"
)

type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]model.Job
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]model.Job)}
}

func (s *JobStore) Create(job model.Job) (model.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return job, nil
}

func (s *JobStore) List() []model.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]model.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		out = append(out, job)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *JobStore) Get(id string) (model.Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	return job, ok
}

func (s *JobStore) Update(job model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jobs[job.ID]; !ok {
		return ErrJobNotFound
	}
	s.jobs[job.ID] = job
	return nil
}

func (s *JobStore) assignable() []model.Job {
	out := make([]model.Job, 0)
	for _, job := range s.jobs {
		if job.Status == model.JobPending || job.Status == model.JobRetrying {
			out = append(out, job)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *JobStore) AssignNext(agentID string, capabilities []model.JobType) (model.Job, bool) {
	capSet := make(map[model.JobType]struct{}, len(capabilities))
	for _, c := range capabilities {
		capSet[c] = struct{}{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.assignable() {
		if _, ok := capSet[job.Type]; !ok {
			continue
		}
		job.Status = model.JobRunning
		job.AssignedAgentID = agentID
		job.AttemptCount++
		job.StartedAt = nil
		job.FinishedAt = nil
		job.Error = ""
		job.Result = nil
		s.jobs[job.ID] = job
		return job, true
	}
	return model.Job{}, false
}

func (s *JobStore) Complete(id string, job model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jobs[id]; !ok {
		return ErrJobNotFound
	}
	s.jobs[id] = job
	return nil
}
