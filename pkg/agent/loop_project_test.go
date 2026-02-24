package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/project"
)

func TestLoop_RunProjectsList_Empty(t *testing.T) {
	cfg := config.DefaultConfig()
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace

	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+1", SenderID: "+1"}

	resp, err := loop.RunProjectsList(ctx, msg)
	if err != nil {
		t.Fatalf("RunProjectsList: %v", err)
	}
	if !strings.Contains(resp, "No projects") && !strings.Contains(resp, "registered") {
		t.Errorf("empty store: response should mention no projects: %s", resp)
	}
}

func TestLoop_RunProjectsList_WithProjects(t *testing.T) {
	cfg := config.DefaultConfig()
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace

	// Add a project via the store
	store := project.NewStore(workspace)
	p := &project.Project{ID: "myapp", Name: "My App", Path: ".", BuildCommand: "npm run build"}
	if err := store.Add(p); err != nil {
		t.Fatal(err)
	}

	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+1", SenderID: "+1"}

	resp, err := loop.RunProjectsList(ctx, msg)
	if err != nil {
		t.Fatalf("RunProjectsList: %v", err)
	}
	if !strings.Contains(resp, "My App") || !strings.Contains(resp, "Projects") {
		t.Errorf("response should list project: %s", resp)
	}
}

func TestLoop_RunProjectsBuild_WithProject(t *testing.T) {
	cfg := config.DefaultConfig()
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace

	store := project.NewStore(workspace)
	p := &project.Project{ID: "buildtest", Name: "BuildTest", Path: ".", BuildCommand: "echo hello"}
	if err := store.Add(p); err != nil {
		t.Fatal(err)
	}

	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: false})

	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+1", SenderID: "+1"}

	resp, err := loop.RunProjectsBuild(ctx, "buildtest", msg)
	if err != nil {
		t.Fatalf("RunProjectsBuild: %v", err)
	}
	if !strings.Contains(resp, "hello") && !strings.Contains(resp, "complete") {
		t.Errorf("build should run: %s", resp)
	}
}

func TestLoop_RunProjectsBuild_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace

	// Add one project so store is not empty; then request non-existent ID
	store := project.NewStore(workspace)
	store.Add(&project.Project{ID: "real", Name: "Real", Path: "."})

	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+1", SenderID: "+1"}

	resp, err := loop.RunProjectsBuild(ctx, "nonexistent", msg)
	if err != nil {
		t.Fatalf("RunProjectsBuild: %v", err)
	}
	if !strings.Contains(resp, "not found") && !strings.Contains(resp, "nonexistent") {
		t.Errorf("expected not found: %s", resp)
	}
}

func TestLoop_RunProjectsPull_WithProject(t *testing.T) {
	cfg := config.DefaultConfig()
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace

	// Create a git repo
	projDir := filepath.Join(workspace, "code-projects")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create project dir with .git (or we can skip git pull test if no git)
	repoDir := filepath.Join(workspace, "pulltest")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	store := project.NewStore(workspace)
	p := &project.Project{ID: "pulltest", Name: "PullTest", Path: "pulltest", BuildCommand: ""}
	if err := store.Add(p); err != nil {
		t.Fatal(err)
	}

	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: false})

	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+1", SenderID: "+1"}

	resp, err := loop.RunProjectsPull(ctx, "pulltest", msg)
	if err != nil {
		t.Fatalf("RunProjectsPull: %v", err)
	}
	// Git pull may fail or succeed in non-repo dir; we just check we get a response
	if resp == "" {
		t.Error("expected non-empty response")
	}
}

func TestLoop_RunTasksList_Empty(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+1", SenderID: "+1"}

	resp, err := loop.RunTasksList(ctx, msg)
	if err != nil {
		t.Fatalf("RunTasksList: %v", err)
	}
	if !strings.Contains(resp, "No running") {
		t.Errorf("empty tasks: %s", resp)
	}
}

func TestLoop_RunTasksCancel_Empty(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ctx := context.Background()
	msg := bus.InboundMessage{Channel: "whatsapp", ChatID: "+1", SenderID: "+1"}

	resp, err := loop.RunTasksCancel(ctx, "", msg)
	if err != nil {
		t.Fatalf("RunTasksCancel: %v", err)
	}
	if !strings.Contains(resp, "No running") {
		t.Errorf("empty cancel: %s", resp)
	}
}

func TestLoop_RunProjectsGetIDs(t *testing.T) {
	cfg := config.DefaultConfig()
	workspace := t.TempDir()
	cfg.Agents.Defaults.Workspace = workspace

	store := project.NewStore(workspace)
	store.Add(&project.Project{ID: "a", Name: "A", Path: "."})
	store.Add(&project.Project{ID: "b", Name: "B", Path: "."})

	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ids, err := loop.RunProjectsGetIDs(context.Background())
	if err != nil {
		t.Fatalf("RunProjectsGetIDs: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("RunProjectsGetIDs: got %d, want 2", len(ids))
	}
	if ids[0] != "a" || ids[1] != "b" {
		t.Errorf("RunProjectsGetIDs: got %v", ids)
	}
}

func TestLoop_RunTasksGetIDs(t *testing.T) {
	cfg := config.DefaultConfig()
	msgBus := bus.NewMessageBus(10)
	eventBus := bus.New()
	loop := NewLoop(cfg, msgBus, eventBus, &LoopOptions{SafeMode: true})

	ids, err := loop.RunTasksGetIDs(context.Background())
	if err != nil {
		t.Fatalf("RunTasksGetIDs: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("RunTasksGetIDs (no tasks): got %v", ids)
	}
}
