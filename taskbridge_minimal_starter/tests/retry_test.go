package tests

import (
	"testing"
	"time"

	"taskbridge/internal/model"
	"taskbridge/internal/scheduler"
	"taskbridge/internal/store"
)

func TestRetryLogic(t *testing.T) {
	job := model.Job{AttemptCount: 1, MaxRetries: 2}
	if scheduler.ResolveCompletionStatus(job, model.JobFailed) != model.JobRetrying {
		t.Fatal("expected RETRYING on first failure with max_retries=2")
	}

	job.AttemptCount = 3
	if scheduler.ResolveCompletionStatus(job, model.JobFailed) != model.JobFailed {
		t.Fatal("expected permanent FAILED when retries exhausted")
	}
}

func TestCompleteJobRetriesThenFails(t *testing.T) {
	st := store.NewMemoryStore()
	_, _ = st.CreateJob(model.Job{
		ID:         "job-1",
		Name:       "fail",
		Type:       model.JobWait,
		Status:     model.JobRunning,
		CreatedAt:  time.Now().UTC(),
		MaxRetries: 1,
		AttemptCount: 1,
	})

	if err := st.CompleteJob("job-1", model.JobFailed, nil, nil, "boom"); err != nil {
		t.Fatal(err)
	}
	job, _, _ := st.GetJob("job-1")
	if job.Status != model.JobRetrying {
		t.Fatalf("expected RETRYING, got %s", job.Status)
	}

	job.Status = model.JobRunning
	job.AttemptCount = 2
	_ = st.UpdateJob(job)
	if err := st.CompleteJob("job-1", model.JobFailed, nil, nil, "boom again"); err != nil {
		t.Fatal(err)
	}
	job, _, _ = st.GetJob("job-1")
	if job.Status != model.JobFailed {
		t.Fatalf("expected FAILED, got %s", job.Status)
	}
}

func TestTimeoutDetection(t *testing.T) {
	started := time.Now().UTC().Add(-20 * time.Second)
	job := model.Job{
		Status:         model.JobRunning,
		StartedAt:      &started,
		TimeoutSeconds: 10,
		AttemptCount:   1,
		MaxRetries:     0,
	}

	if !scheduler.IsTimedOut(job, time.Now().UTC()) {
		t.Fatal("expected job to be timed out")
	}

	updated := scheduler.ApplyTimeout(job, time.Now().UTC())
	if updated.Status != model.JobFailed {
		t.Fatalf("expected FAILED after timeout, got %s", updated.Status)
	}
}
