package executor

import (
	"context"
	"taskbridge/internal/model"
)

type Result struct {
	Status model.JobStatus `json:"status"`
	Logs   []string        `json:"logs"`
	Result map[string]any  `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type Executor interface {
	Type() model.JobType
	Execute(ctx context.Context, job model.Job) Result
}

type Registry struct {
	executors map[model.JobType]Executor
}

func NewRegistry() *Registry {
	return &Registry{executors: map[model.JobType]Executor{}}
}

func (r *Registry) Register(ex Executor) {
	r.executors[ex.Type()] = ex
}

func (r *Registry) Get(t model.JobType) (Executor, bool) {
	ex, ok := r.executors[t]
	return ex, ok
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	for _, ex := range []Executor{
		HTTPCheckExecutor{},
		TCPCheckExecutor{},
		FileExistsExecutor{},
		ChecksumExecutor{},
		CopyFileExecutor{},
		WriteFileExecutor{},
		WaitExecutor{},
	} {
		r.Register(ex)
	}
	return r
}
