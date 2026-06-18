package model

type CreateJobRequest struct {
	Name           string         `json:"name"`
	Type           JobType        `json:"type"`
	Payload        map[string]any `json:"payload"`
	MaxRetries     int            `json:"max_retries"`
	TimeoutSeconds int            `json:"timeout_seconds"`
}

type RegisterAgentRequest struct {
	Hostname     string    `json:"hostname"`
	Capabilities []JobType `json:"capabilities"`
}

type RegisterAgentResponse struct {
	Agent Agent `json:"agent"`
}

type JobResultRequest struct {
	Status JobStatus      `json:"status"`
	Logs   []string       `json:"logs"`
	Result map[string]any `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}

type NextJobResponse struct {
	Job   *Job `json:"job"`
	Found bool `json:"found"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
