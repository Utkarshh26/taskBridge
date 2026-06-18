package executor

import "taskbridge/internal/model"

func successResult(logs []string, result map[string]any) Result {
	return Result{
		Status: model.JobSuccess,
		Logs:   logs,
		Result: result,
	}
}

func failResult(msg string) Result {
	return Result{
		Status: model.JobFailed,
		Logs:   []string{msg},
		Error:  msg,
	}
}
