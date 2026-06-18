package scheduler

import (
	"time"

	"taskbridge/internal/model"
)

func ResolveCompletionStatus(job model.Job, reported model.JobStatus) model.JobStatus {
	if reported == model.JobSuccess {
		return model.JobSuccess
	}

	if job.AttemptCount <= job.MaxRetries {
		return model.JobRetrying
	}
	return model.JobFailed
}

func IsTimedOut(job model.Job, now time.Time) bool {
	if job.Status != model.JobRunning || job.StartedAt == nil {
		return false
	}
	if job.TimeoutSeconds <= 0 {
		return false
	}
	deadline := job.StartedAt.Add(time.Duration(job.TimeoutSeconds) * time.Second)
	return now.After(deadline)
}

func ApplyTimeout(job model.Job, now time.Time) model.Job {
	job.Logs = append(job.Logs, "job timed out")
	job.Error = "execution exceeded timeout"
	job.FinishedAt = &now
	job.AssignedAgentID = ""
	job.Status = ResolveCompletionStatus(job, model.JobFailed)
	return job
}
