package idempotency

import (
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	c := New(100 * time.Millisecond)

	// Initially empty
	if _, _, ok := c.Get("session1", "hello"); ok {
		t.Error("Get on empty cache should return ok=false")
	}

	// Set and Get
	c.Set("session1", "hello", "task-1", "result-1")
	taskID, result, ok := c.Get("session1", "hello")
	if !ok {
		t.Fatal("Get after Set should return ok=true")
	}
	if taskID != "task-1" {
		t.Errorf("taskID = %q, want task-1", taskID)
	}
	if result != "result-1" {
		t.Errorf("result = %q, want result-1", result)
	}

	// Different content misses
	if _, _, ok := c.Get("session1", "world"); ok {
		t.Error("different content should miss")
	}

	// Different session misses
	if _, _, ok := c.Get("session2", "hello"); ok {
		t.Error("different session should miss")
	}
}

func TestCache_TTLExpiry(t *testing.T) {
	c := New(50 * time.Millisecond)
	c.Set("s", "c", "t", "r")

	if _, _, ok := c.Get("s", "c"); !ok {
		t.Fatal("should hit before TTL")
	}

	time.Sleep(60 * time.Millisecond)
	if _, _, ok := c.Get("s", "c"); ok {
		t.Error("should miss after TTL")
	}
}

func TestCache_Cleanup(t *testing.T) {
	c := New(10 * time.Millisecond)
	c.Set("s1", "c1", "t1", "r1")
	c.Set("s2", "c2", "t2", "r2")

	time.Sleep(20 * time.Millisecond)
	c.Cleanup()

	if _, _, ok := c.Get("s1", "c1"); ok {
		t.Error("Cleanup should remove expired entries")
	}
	if _, _, ok := c.Get("s2", "c2"); ok {
		t.Error("Cleanup should remove expired entries")
	}
}
