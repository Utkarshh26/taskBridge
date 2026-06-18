# TaskBridge

TaskBridge is a small distributed job runner built in Go. The project consists of a central server that manages jobs and one or more agents that poll for work, execute supported tasks, and send results back to the server.

This project was built as part of a Golang internship assignment to explore REST APIs, concurrency, job scheduling, and communication between services.

## Project Structure

```text
cmd/
├── agent
└── server

internal/
├── agent
├── executor
├── model
├── scheduler
├── server
└── store

examples/
```

## Features

* Create and manage jobs through REST APIs
* Agent registration and heartbeat tracking
* Job polling and execution
* Retry support for failed jobs
* Multiple job types
* Thread-safe in-memory storage
* Unit tests

Supported job types:

* http_check
* tcp_check
* file_exists
* checksum
* copy_file
* write_file
* wait

## Prerequisites

* Go 1.22 or later

## Running the Project

Make sure you are inside the project directory:

```bash
cd taskbridge_minimal_starter
```

### Start the Server

```bash
go run ./cmd/server --addr :8080
```

Expected output:

```text
TaskBridge server listening on :8080
```

Verify the server:

```bash
curl http://localhost:8080/health
```

### Start an Agent

Open another terminal:

```bash
go run ./cmd/agent --server http://localhost:8080
```

The agent automatically:

* Registers with the server
* Sends heartbeats every 10 seconds
* Polls for jobs every 3 seconds

### Create a Job

Open a third terminal:

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d @examples/create-wait-job.json
```

View jobs:

```bash
curl http://localhost:8080/jobs
```

## Example Workflow

1. Start the server
2. Start an agent
3. Submit a job
4. Watch the job move from PENDING to SUCCESS

Check job status:

```bash
curl http://localhost:8080/jobs
```

## Available Flags

### Server

```bash
--addr :8080
```

Server listen address.

### Agent

```bash
--server http://localhost:8080
```

Server URL.

```bash
--capabilities http_check,tcp_check,file_exists,checksum,copy_file,write_file,wait
```

Supported job types.

```bash
--poll-interval 3s
```

Job polling interval.

```bash
--heartbeat-interval 10s
```

Heartbeat interval.

## Running Tests

```bash
go test ./...
```

Run with race detection:

```bash
go test -race ./...
```

## Notes

This implementation uses in-memory storage, so all jobs and agent information are lost when the server stops.

The focus of the project is demonstrating job execution, retries, agent communication, and concurrency handling in Go.
