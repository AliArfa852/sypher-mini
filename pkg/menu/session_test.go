package menu

import (
	"testing"
	"time"
)

func TestKey(t *testing.T) {
	k := Key("whatsapp", "+123")
	if k != "whatsapp:+123" {
		t.Errorf("Key = %q, want whatsapp:+123", k)
	}
}

func TestSessionStore_GetSet(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := "whatsapp:+123"

	_, ok := store.Get(key)
	if ok {
		t.Error("Get on empty store should return false")
	}

	store.Set(key, "cli", "main")
	ses, ok := store.Get(key)
	if !ok || ses == nil {
		t.Fatal("Get should return session after Set")
	}
	if ses.CurrentMenu != "cli" || ses.ParentMenu != "main" {
		t.Errorf("session = %+v", ses)
	}
}

func TestSessionStore_Clear(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := "whatsapp:+123"
	store.Set(key, "cli", "main")
	store.Clear(key)
	_, ok := store.Get(key)
	if ok {
		t.Error("Get after Clear should return false")
	}
}

func TestSessionStore_ResetToMain(t *testing.T) {
	store := NewSessionStore(10 * time.Minute)
	key := "whatsapp:+123"
	store.Set(key, "cli", "main")
	store.ResetToMain(key)
	ses, ok := store.Get(key)
	if !ok || ses.CurrentMenu != "main" || ses.ParentMenu != "" {
		t.Errorf("ResetToMain: got %+v", ses)
	}
}

func TestSessionStore_TTL(t *testing.T) {
	store := NewSessionStore(100 * time.Millisecond)
	key := "whatsapp:+123"
	store.Set(key, "main", "")
	_, ok := store.Get(key)
	if !ok {
		t.Error("Get immediately after Set should succeed")
	}
	time.Sleep(150 * time.Millisecond)
	_, ok = store.Get(key)
	if ok {
		t.Error("Get after TTL should return false")
	}
}
