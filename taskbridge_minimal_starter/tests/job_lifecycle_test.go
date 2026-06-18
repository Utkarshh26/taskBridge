package tests

import (
	"testing"
	"time"

	"taskbridge/internal/model"
	"taskbridge/internal/store"
)

func TestJobLifecycleSuccess(t *testing.T) {
	st := store.NewMemoryStore()

	_, _ = st.RegisterAgent(model.Agent{
		ID:           "agent-1",
		Hostname:     "host-a",
		Capabilities: []model.JobType{model.JobWait},
		LastSeen:     time.Now().UTC(),
		Status:       model.AgentStatusOnline,
	})

	_, _ = st.CreateJob(model.Job{
		ID:             "job-1",
		Name:           "wait",
		Type:           model.JobWait,
		Status:         model.JobPending,
		CreatedAt:      time.Now().UTC(),
		MaxRetries:     2,
		TimeoutSeconds: 10,
	})

	assigned, found, err := st.AssignNextJob("agent-1", []model.JobType{model.JobWait})
	if err != nil || !found {
		t.Fatal("expected job assignment")
	}

	err = st.CompleteJob(assigned.ID, model.JobSuccess, []string{"done"}, map[string]any{"ok": true}, "")
	if err != nil {
		t.Fatal(err)
	}

	job, ok, _ := st.GetJob(assigned.ID)
	if !ok || job.Status != model.JobSuccess {
		t.Fatalf("expected SUCCESS, got %s", job.Status)
	}
	if job.FinishedAt == nil {
		t.Fatal("expected finished_at to be set")
	}
}
