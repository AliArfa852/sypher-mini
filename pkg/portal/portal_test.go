package portal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/agent"
	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestPortal_GET_ApiTasks(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: true})

	mux := http.NewServeMux()
	Register(mux, cfg, loop, msgBus)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/tasks status = %d, want 200", rec.Code)
	}
	var resp struct {
		Tasks []interface{} `json:"tasks"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Tasks == nil {
		t.Error("expected tasks array")
	}
}

func TestPortal_GET_ApiConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: true})

	mux := http.NewServeMux()
	Register(mux, cfg, loop, msgBus)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/config status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["agents"] == nil {
		t.Error("expected agents in config")
	}
}

func TestPortal_POST_ApiTasks(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: true})

	mux := http.NewServeMux()
	Register(mux, cfg, loop, msgBus)

	body := bytes.NewBufferString(`{"content":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/tasks status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["ok"] != "true" && resp["ok"] != true {
		t.Errorf("expected ok=true, got %v", resp["ok"])
	}
}

func TestPortal_POST_ApiTasks_EmptyContent(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: true})

	mux := http.NewServeMux()
	Register(mux, cfg, loop, msgBus)

	body := bytes.NewBufferString(`{"content":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /api/tasks with empty content status = %d, want 400", rec.Code)
	}
}

func TestPortal_GET_ApiAgents(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: true})

	mux := http.NewServeMux()
	Register(mux, cfg, loop, msgBus)

	req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/agents status = %d, want 200", rec.Code)
	}
	var resp struct {
		Agents []interface{} `json:"agents"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Agents == nil {
		t.Error("expected agents array")
	}
}

func TestPortal_GET_ApiTasksId_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: true})

	mux := http.NewServeMux()
	Register(mux, cfg, loop, msgBus)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/nonexistent-id", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/tasks/nonexistent status = %d, want 404", rec.Code)
	}
}

func TestPortal_POST_ApiTasksCancel(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := agent.NewLoop(cfg, msgBus, eventBus, &agent.LoopOptions{SafeMode: true})

	mux := http.NewServeMux()
	Register(mux, cfg, loop, msgBus)

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/fake-id/cancel", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/tasks/id/cancel status = %d, want 200", rec.Code)
	}
	var resp map[string]bool
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["ok"] {
		t.Error("expected ok=false for non-existent task")
	}
}
