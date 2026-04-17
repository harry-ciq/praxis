package worker

import (
	"encoding/json"
	"testing"
)

func TestNewProviderSyncTask(t *testing.T) {
	task, err := NewProviderSyncTask("user-123", "GITHUB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type() != TaskProviderSync {
		t.Errorf("expected type %q, got %q", TaskProviderSync, task.Type())
	}

	var p ProviderSyncPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if p.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got %q", p.UserID)
	}
	if p.Provider != "GITHUB" {
		t.Errorf("expected Provider 'GITHUB', got %q", p.Provider)
	}
}

func TestNewEmailNotifyTask(t *testing.T) {
	data := map[string]interface{}{"name": "Harry"}
	task, err := NewEmailNotifyTask("foo@bar.com", "Hello", "welcome", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type() != TaskEmailNotify {
		t.Errorf("expected type %q, got %q", TaskEmailNotify, task.Type())
	}

	var p EmailNotifyPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if p.To != "foo@bar.com" || p.Subject != "Hello" || p.Template != "welcome" {
		t.Errorf("unexpected payload: %+v", p)
	}
	if p.Data["name"] != "Harry" {
		t.Errorf("expected data.name 'Harry', got %v", p.Data["name"])
	}
}
