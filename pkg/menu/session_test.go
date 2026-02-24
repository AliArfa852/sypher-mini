package menu

import (
	"testing"
	"time"
)

func TestSessionStore_Key(t *testing.T) {
	k := Key("whatsapp", "+123")
	if k != "whatsapp:+123" {
		t.Errorf("Key = %q, want whatsapp:+123", k)
	}
}

func TestSessionStore_Set_Get(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := Key("whatsapp", "chat1")
	store.Set(key, "projects", "main")

	ses, ok := store.Get(key)
	if !ok || ses == nil {
		t.Fatal("Get should return session")
	}
	if ses.CurrentMenu != "projects" || ses.ParentMenu != "main" {
		t.Errorf("session = %+v", ses)
	}
}

func TestSessionStore_Get_Expired(t *testing.T) {
	store := NewSessionStore(1 * time.Millisecond)
	key := Key("whatsapp", "chat1")
	store.Set(key, "main", "")
	time.Sleep(5 * time.Millisecond)

	_, ok := store.Get(key)
	if ok {
		t.Error("expired session should not be returned")
	}
}

func TestSessionStore_Clear(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := Key("whatsapp", "chat1")
	store.Set(key, "main", "")
	store.Clear(key)

	_, ok := store.Get(key)
	if ok {
		t.Error("Clear should remove session")
	}
}

func TestSessionStore_ResetToMain(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := Key("whatsapp", "chat1")
	store.Set(key, "projects", "main")
	store.ResetToMain(key)

	ses, _ := store.Get(key)
	if ses.CurrentMenu != "main" || ses.ParentMenu != "" {
		t.Errorf("ResetToMain: got %+v", ses)
	}
}

func TestSessionStore_SetPendingProjectAction(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := Key("whatsapp", "chat1")
	store.Set(key, "projects", "main")
	store.SetPendingProjectAction(key, "projects_build", []string{"p1", "p2"})

	action, projectIDs, taskIDs, ok := store.GetPendingProjectAction(key)
	if !ok {
		t.Fatal("GetPendingProjectAction should return true")
	}
	if action != "projects_build" {
		t.Errorf("action = %q, want projects_build", action)
	}
	if len(projectIDs) != 2 || projectIDs[0] != "p1" {
		t.Errorf("projectIDs = %v", projectIDs)
	}
	if len(taskIDs) != 0 {
		t.Errorf("taskIDs should be empty, got %v", taskIDs)
	}
}

func TestSessionStore_SetPendingTaskCancel(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := Key("whatsapp", "chat1")
	store.Set(key, "tasks", "main")
	store.SetPendingTaskCancel(key, []string{"t1", "t2"})

	action, projectIDs, taskIDs, ok := store.GetPendingProjectAction(key)
	if !ok {
		t.Fatal("GetPendingProjectAction should return true")
	}
	if action != "tasks_cancel" {
		t.Errorf("action = %q, want tasks_cancel", action)
	}
	if len(taskIDs) != 2 || taskIDs[0] != "t1" {
		t.Errorf("taskIDs = %v", taskIDs)
	}
	if len(projectIDs) != 0 {
		t.Errorf("projectIDs should be empty, got %v", projectIDs)
	}
}

func TestSessionStore_ClearPendingProjectAction(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := Key("whatsapp", "chat1")
	store.Set(key, "projects", "main")
	store.SetPendingProjectAction(key, "projects_build", []string{"p1"})
	store.ClearPendingProjectAction(key)

	_, _, _, ok := store.GetPendingProjectAction(key)
	if ok {
		t.Error("ClearPendingProjectAction should clear state")
	}
}

func TestSessionStore_SetPending_CreatesSession(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := Key("whatsapp", "newchat")
	store.SetPendingProjectAction(key, "projects_build", []string{"p1"})

	action, ids, _, ok := store.GetPendingProjectAction(key)
	if !ok {
		t.Fatal("SetPendingProjectAction should create session")
	}
	if action != "projects_build" || len(ids) != 1 {
		t.Errorf("action=%q ids=%v", action, ids)
	}
}
