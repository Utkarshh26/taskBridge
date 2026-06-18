package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"taskbridge/internal/model"
	"taskbridge/internal/server"
	"taskbridge/internal/store"
)

func TestHealthHandler(t *testing.T) {
	srv := server.New(store.NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateAndGetJobHandler(t *testing.T) {
	srv := server.New(store.NewMemoryStore())

	body := model.CreateJobRequest{
		Name:           "wait-job",
		Type:           model.JobWait,
		Payload:        map[string]any{"duration_seconds": 1},
		MaxRetries:     1,
		TimeoutSeconds: 5,
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var created model.Job
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest(http.MethodGet, "/jobs/"+created.ID, nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRegisterAgentAndAssignJob(t *testing.T) {
	srv := server.New(store.NewMemoryStore())

	regBody := model.RegisterAgentRequest{
		Hostname:     "test-host",
		Capabilities: []model.JobType{model.JobWait},
	}
	data, _ := json.Marshal(regBody)

	req := httptest.NewRequest(http.MethodPost, "/agents/register", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var regResp model.RegisterAgentResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &regResp); err != nil {
		t.Fatal(err)
	}

	jobBody := model.CreateJobRequest{
		Name:           "wait",
		Type:           model.JobWait,
		Payload:        map[string]any{"duration_seconds": 1},
		TimeoutSeconds: 5,
	}
	data, _ = json.Marshal(jobBody)
	req = httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(data))
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected job create 201, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/agents/"+regResp.Agent.ID+"/next-job", nil)
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var next model.NextJobResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &next); err != nil {
		t.Fatal(err)
	}
	if !next.Found || next.Job == nil {
		t.Fatal("expected job to be assigned")
	}
}
