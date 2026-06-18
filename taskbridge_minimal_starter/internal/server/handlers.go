package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"taskbridge/internal/model"
	"taskbridge/internal/scheduler"
	"taskbridge/internal/store"
)

type Server struct {
	store store.Store
}

func New(st store.Store) *Server {
	return &Server{store: st}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/jobs", s.handleJobs)
	mux.HandleFunc("/jobs/", s.handleJobByID)
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/agents/register", s.handleRegisterAgent)
	mux.HandleFunc("/agents/", s.handleAgentByID)
	return mux
}

func (s *Server) RunTimeoutWatcher(stop <-chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			s.expireTimedOutJobs()
		}
	}
}

func (s *Server) expireTimedOutJobs() {
	jobs, err := s.store.ListJobs()
	if err != nil {
		return
	}
	now := time.Now().UTC()
	for _, job := range jobs {
		if !scheduler.IsTimedOut(job, now) {
			continue
		}
		updated := scheduler.ApplyTimeout(job, now)
		_ = s.store.UpdateJob(updated)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "taskbridge-server",
	})
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createJob(w, r)
	case http.MethodGet:
		s.listJobs(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	var req model.CreateJobRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 30
	}

	job := model.Job{
		ID:             uuid.NewString(),
		Name:           req.Name,
		Type:           req.Type,
		Payload:        req.Payload,
		Status:         model.JobPending,
		CreatedAt:      time.Now().UTC(),
		MaxRetries:     req.MaxRetries,
		TimeoutSeconds: req.TimeoutSeconds,
	}
	if job.Payload == nil {
		job.Payload = map[string]any{}
	}

	created, err := s.store.CreateJob(job)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.store.ListJobs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if jobs == nil {
		jobs = []model.Job{}
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) handleJobByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/jobs/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	jobID := parts[0]

	if len(parts) == 2 && parts[1] == "result" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.submitJobResult(w, r, jobID)
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	job, ok, err := s.store.GetJob(jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) submitJobResult(w http.ResponseWriter, r *http.Request, jobID string) {
	var req model.JobResultRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.store.CompleteJob(jobID, req.Status, req.Logs, req.Result, req.Error); err != nil {
		if err == store.ErrJobNotFound {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	job, ok, _ := s.store.GetJob(jobID)
	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	agents, err := s.store.ListAgents()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if agents == nil {
		agents = []model.Agent{}
	}
	writeJSON(w, http.StatusOK, agents)
}

func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.RegisterAgentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Hostname == "" {
		writeError(w, http.StatusBadRequest, "hostname is required")
		return
	}

	agent := model.Agent{
		ID:           uuid.NewString(),
		Hostname:     req.Hostname,
		Capabilities: req.Capabilities,
		LastSeen:     time.Now().UTC(),
		Status:       model.AgentStatusOnline,
	}

	created, err := s.store.RegisterAgent(agent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, model.RegisterAgentResponse{Agent: created})
}

func (s *Server) handleAgentByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/agents/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	agentID := parts[0]
	action := parts[1]

	switch action {
	case "heartbeat":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := s.store.Heartbeat(agentID); err != nil {
			if err == store.ErrAgentNotFound {
				writeError(w, http.StatusNotFound, "agent not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case "next-job":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.assignNextJob(w, r, agentID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s *Server) assignNextJob(w http.ResponseWriter, r *http.Request, agentID string) {
	agent, ok, err := s.store.GetAgent(agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}

	job, found, err := s.store.AssignNextJob(agentID, agent.Capabilities)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := model.NextJobResponse{Found: found}
	if found {
		resp.Job = &job
	}
	writeJSON(w, http.StatusOK, resp)
}
