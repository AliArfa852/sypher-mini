package logging

import (
	"testing"
)

func TestNew_Stdout(t *testing.T) {
	l, err := New(Config{Output: "stdout", JSON: true})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if l == nil {
		t.Fatal("New returned nil")
	}
}

func TestLogger_Log_JSONDisabled(t *testing.T) {
	l, _ := New(Config{Output: "stdout", JSON: false})
	// Should not panic
	l.Log("info", "test", nil)
}

func TestLogger_Info(t *testing.T) {
	l, _ := New(Config{Output: "stdout", JSON: true})
	l.Info("hello", map[string]interface{}{"key": "value"})
	// No panic
}

func TestLogger_Error(t *testing.T) {
	l, _ := New(Config{Output: "stdout", JSON: true})
	l.Error("oops", nil)
	// No panic
}

func TestLogger_Log_WithFields(t *testing.T) {
	l, _ := New(Config{Output: "stdout", JSON: true})
	l.Log("info", "msg", map[string]interface{}{
		"string": "v",
		"int":    42,
		"nested": map[string]string{"a": "b"},
	})
}
