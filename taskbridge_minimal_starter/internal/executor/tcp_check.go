package executor

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"taskbridge/internal/model"
)

type TCPCheckExecutor struct{}

func (TCPCheckExecutor) Type() model.JobType { return model.JobTCPCheck }

func (TCPCheckExecutor) Execute(ctx context.Context, job model.Job) Result {
	host, _ := job.Payload["host"].(string)
	if host == "" {
		return failResult("host is required")
	}

	port := 80
	if v, ok := job.Payload["port"]; ok {
		switch n := v.(type) {
		case float64:
			port = int(n)
		case int:
			port = n
		case string:
			parsed, err := strconv.Atoi(n)
			if err != nil {
				return failResult("invalid port")
			}
			port = parsed
		}
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return failResult(err.Error())
	}
	conn.Close()

	return successResult([]string{fmt.Sprintf("connected to %s", addr)}, map[string]any{
		"address": addr,
	})
}
