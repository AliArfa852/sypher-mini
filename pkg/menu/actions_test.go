package menu

import (
	"context"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
)

func TestExecuteAction_Help(t *testing.T) {
	resp, err := ExecuteAction(context.Background(), "help", config.DefaultConfig(), nil, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp, "Menu") || !strings.Contains(resp, "sypher") {
		t.Errorf("help response: %s", resp)
	}
}

func TestExecuteAction_AddAPI(t *testing.T) {
	resp, err := ExecuteAction(context.Background(), "add_api", config.DefaultConfig(), nil, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp, "API") || !strings.Contains(resp, "config") {
		t.Errorf("add_api response: %s", resp)
	}
}

func TestExecuteAction_ConnectGemini(t *testing.T) {
	resp, err := ExecuteAction(context.Background(), "connect_gemini", config.DefaultConfig(), nil, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp, "Gemini") {
		t.Errorf("connect_gemini response: %s", resp)
	}
}

func TestRenderMenu(t *testing.T) {
	cfg := DefaultMenus()
	resp := RenderMenu(cfg, "main")
	if !strings.Contains(resp, "Control Panel") || !strings.Contains(resp, "Projects") {
		t.Errorf("RenderMenu main: %s", resp)
	}
	if !strings.Contains(resp, "Roll the dice") {
		t.Errorf("RenderMenu main missing dice: %s", resp)
	}
}

func TestExecuteAction_RollDice(t *testing.T) {
	for _, action := range []string{"roll_1d6", "roll_2d6", "roll_1d20"} {
		resp, err := ExecuteAction(context.Background(), action, config.DefaultConfig(), nil, bus.InboundMessage{})
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if !strings.Contains(resp, "🎲") || !strings.Contains(resp, "rolled") {
			t.Errorf("%s response: %s", action, resp)
		}
	}
}

func TestExecuteAction_ProjectsList_WithRunner(t *testing.T) {
	runner := &mockActionRunner{projectsList: "1. proj1\n2. proj2"}
	resp, err := ExecuteAction(context.Background(), "projects_list", config.DefaultConfig(), runner, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if resp != "1. proj1\n2. proj2" {
		t.Errorf("projects_list with runner: %s", resp)
	}
}

func TestExecuteAction_TasksList_WithRunner(t *testing.T) {
	runner := &mockActionRunner{tasksList: "No running tasks"}
	resp, err := ExecuteAction(context.Background(), "tasks_list", config.DefaultConfig(), runner, bus.InboundMessage{})
	if err != nil {
		t.Fatal(err)
	}
	if resp != "No running tasks" {
		t.Errorf("tasks_list with runner: %s", resp)
	}
}

// mockActionRunner implements ActionRunner for actions_test.
type mockActionRunner struct {
	projectsList string
	tasksList    string
}

func (m *mockActionRunner) RunCliList(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return "", nil
}
func (m *mockActionRunner) RunCliNew(ctx context.Context, tag string, msg bus.InboundMessage) (string, error) {
	return "", nil
}
func (m *mockActionRunner) RunStatus(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return "", nil
}
func (m *mockActionRunner) RunConfigStatus(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return "", nil
}
func (m *mockActionRunner) RunProjectsList(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return m.projectsList, nil
}
func (m *mockActionRunner) RunProjectsBuild(ctx context.Context, projectID string, msg bus.InboundMessage) (string, error) {
	return "", nil
}
func (m *mockActionRunner) RunProjectsPull(ctx context.Context, projectID string, msg bus.InboundMessage) (string, error) {
	return "", nil
}
func (m *mockActionRunner) RunProjectsGetIDs(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (m *mockActionRunner) RunTasksList(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return m.tasksList, nil
}
func (m *mockActionRunner) RunTasksCancel(ctx context.Context, taskID string, msg bus.InboundMessage) (string, error) {
	return "", nil
}
func (m *mockActionRunner) RunTasksGetIDs(ctx context.Context) ([]string, error) {
	return nil, nil
}
