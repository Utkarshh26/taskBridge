package store

import "errors"

var (
	ErrJobNotFound   = errors.New("job not found")
	ErrAgentNotFound = errors.New("agent not found")
)
