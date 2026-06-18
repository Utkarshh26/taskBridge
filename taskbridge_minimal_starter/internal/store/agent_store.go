package store

import (
	"sort"
	"sync"

	"taskbridge/internal/model"
)

type AgentStore struct {
	mu     sync.RWMutex
	agents map[string]model.Agent
}

func NewAgentStore() *AgentStore {
	return &AgentStore{agents: make(map[string]model.Agent)}
}

func (s *AgentStore) Register(agent model.Agent) (model.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[agent.ID] = agent
	return agent, nil
}

func (s *AgentStore) Update(agent model.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[agent.ID]; !ok {
		return ErrAgentNotFound
	}
	s.agents[agent.ID] = agent
	return nil
}

func (s *AgentStore) List() []model.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]model.Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		out = append(out, agent)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (s *AgentStore) Get(id string) (model.Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, ok := s.agents[id]
	return agent, ok
}
