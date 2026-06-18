package tests

import (
	"testing"
	"time"

	"taskbridge/internal/model"
	"taskbridge/internal/store"
)

func TestMemoryStoreCreateAndGetJob(t *testing.T) {
	st := store.NewMemoryStore()
	job := model.Job{
		ID:        "job-1",
		Name:      "test",
		Type:      model.JobWait,
		Status:    model.JobPending,
		CreatedAt: time.Now().UTC(),
	}

	created, err := st.CreateJob(job)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != job.ID {
		t.Fatalf("expected id %s, got %s", job.ID, created.ID)
	}

	got, ok, err := st.GetJob(job.ID)
	if err != nil || !ok {
		t.Fatalf("job not found: ok=%v err=%v", ok, err)
	}
	if got.Name != job.Name {
		t.Fatalf("expected name %s, got %s", job.Name, got.Name)
	}
}

func TestMemoryStoreRegisterAgent(t *testing.T) {
	st := store.NewMemoryStore()
	agent := model.Agent{
		ID:       "agent-1",
		Hostname: "host-a",
		Status:   model.AgentStatusOnline,
		LastSeen: time.Now().UTC(),
	}

	created, err := st.RegisterAgent(agent)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != agent.ID {
		t.Fatalf("unexpected agent id")
	}

	agents, err := st.ListAgents()
	if err != nil || len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d err=%v", len(agents), err)
	}
}

func TestAssignNextJobRespectsCapabilities(t *testing.T) {
	st := store.NewMemoryStore()

	_, _ = st.RegisterAgent(model.Agent{
		ID:           "agent-1",
		Hostname:     "host-a",
		Capabilities: []model.JobType{model.JobWait},
		LastSeen:     time.Now().UTC(),
		Status:       model.AgentStatusOnline,
	})

	_, _ = st.CreateJob(model.Job{
		ID:        "job-http",
		Name:      "http",
		Type:      model.JobHTTPCheck,
		Status:    model.JobPending,
		CreatedAt: time.Now().UTC(),
	})
	_, _ = st.CreateJob(model.Job{
		ID:        "job-wait",
		Name:      "wait",
		Type:      model.JobWait,
		Status:    model.JobPending,
		CreatedAt: time.Now().UTC().Add(time.Second),
	})

	job, found, err := st.AssignNextJob("agent-1", []model.JobType{model.JobWait})
	if err != nil || !found {
		t.Fatalf("expected wait job assignment, found=%v err=%v", found, err)
	}
	if job.ID != "job-wait" {
		t.Fatalf("expected wait job, got %s", job.ID)
	}
	if job.Status != model.JobRunning {
		t.Fatalf("expected RUNNING, got %s", job.Status)
	}
	if job.AttemptCount != 1 {
		t.Fatalf("expected attempt count 1, got %d", job.AttemptCount)
	}
}
