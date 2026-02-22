package menu

import (
	"context"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

// mockRunner implements ActionRunner for testing pending flows.
type mockRunner struct {
	projectsListResp   string
	projectsBuildResp  string
	projectsPullResp   string
	projectsGetIDs     []string
	tasksListResp      string
	tasksCancelResp    string
	tasksGetIDs        []string
	buildCalledWithID  string
	pullCalledWithID   string
	cancelCalledWithID string
}

func (m *mockRunner) RunCliList(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return "No sessions", nil
}
func (m *mockRunner) RunCliNew(ctx context.Context, tag string, msg bus.InboundMessage) (string, error) {
	return "Created", nil
}
func (m *mockRunner) RunStatus(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return "OK", nil
}
func (m *mockRunner) RunConfigStatus(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return "Config", nil
}
func (m *mockRunner) RunProjectsList(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return m.projectsListResp, nil
}
func (m *mockRunner) RunProjectsBuild(ctx context.Context, projectID string, msg bus.InboundMessage) (string, error) {
	m.buildCalledWithID = projectID
	return m.projectsBuildResp, nil
}
func (m *mockRunner) RunProjectsPull(ctx context.Context, projectID string, msg bus.InboundMessage) (string, error) {
	m.pullCalledWithID = projectID
	return m.projectsPullResp, nil
}
func (m *mockRunner) RunProjectsGetIDs(ctx context.Context) ([]string, error) {
	return m.projectsGetIDs, nil
}
func (m *mockRunner) RunTasksList(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return m.tasksListResp, nil
}
func (m *mockRunner) RunTasksCancel(ctx context.Context, taskID string, msg bus.InboundMessage) (string, error) {
	m.cancelCalledWithID = taskID
	return m.tasksCancelResp, nil
}
func (m *mockRunner) RunTasksGetIDs(ctx context.Context) ([]string, error) {
	return m.tasksGetIDs, nil
}

func TestHandler_PendingProjectBuild_Flow(t *testing.T) {
	runner := &mockRunner{
		projectsListResp:  "1. proj1\n2. proj2",
		projectsBuildResp: "Select project",
		projectsGetIDs:    []string{"proj1", "proj2"},
	}
	h := NewHandler(config.DefaultConfig(), runner, "")
	ctx := context.Background()

	// Navigate to projects menu
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "menu"})
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "1"})

	// Trigger projects_build (item 4 in projects menu)
	handled, resp := h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "4"})
	if !handled {
		t.Fatal("projects_build should be handled")
	}
	if !strings.Contains(resp, "Select") && !strings.Contains(resp, "proj") {
		t.Errorf("response should show project list: %s", resp)
	}

	// Reply with 1 to select first project
	handled2, resp2 := h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+123", SenderID: "+123", Content: "1"})
	if !handled2 {
		t.Fatal("numeric selection should be handled")
	}
	if runner.buildCalledWithID != "proj1" {
		t.Errorf("RunProjectsBuild called with %q, want proj1", runner.buildCalledWithID)
	}
	if resp2 == "" {
		t.Error("expected response from build")
	}
}

func TestHandler_PendingTaskCancel_Flow(t *testing.T) {
	runner := &mockRunner{
		tasksListResp:   "1. task-1\n2. task-2",
		tasksCancelResp: "Cancelled",
		tasksGetIDs:     []string{"task-1", "task-2"},
	}
	h := NewHandler(config.DefaultConfig(), runner, "")
	ctx := context.Background()

	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+456", SenderID: "+456", Content: "menu"})
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+456", SenderID: "+456", Content: "2"}) // tasks

	// Trigger tasks_cancel (item 4)
	handled, _ := h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+456", SenderID: "+456", Content: "4"})
	if !handled {
		t.Fatal("tasks_cancel should be handled")
	}

	// Reply with 2 to select second task
	handled2, _ := h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+456", SenderID: "+456", Content: "2"})
	if !handled2 {
		t.Fatal("numeric selection should be handled")
	}
	if runner.cancelCalledWithID != "task-2" {
		t.Errorf("RunTasksCancel called with %q, want task-2", runner.cancelCalledWithID)
	}
}

func TestHandler_BackClearsPending(t *testing.T) {
	runner := &mockRunner{projectsGetIDs: []string{"p1"}}
	h := NewHandler(config.DefaultConfig(), runner, "")
	ctx := context.Background()

	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+789", SenderID: "+789", Content: "menu"})
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+789", SenderID: "+789", Content: "1"})  // projects
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+789", SenderID: "+789", Content: "4"})  // build - sets pending
	h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+789", SenderID: "+789", Content: "0"})  // back

	// After back, sending 1 should NOT trigger build - should go to projects submenu or main
	handled, resp := h.Handle(ctx, bus.InboundMessage{Channel: "whatsapp", ChatID: "+789", SenderID: "+789", Content: "1"})
	if !handled {
		t.Fatal("should be handled")
	}
	// Back goes to parent (main), so 1 would be main menu item 1 = Projects
	if !strings.Contains(resp, "Projects") && !strings.Contains(resp, "Control Panel") {
		t.Errorf("after back, 1 should show menu: %s", resp)
	}
}
