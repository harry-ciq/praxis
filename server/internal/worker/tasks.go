package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task type constants used as the asynq task type string.
const (
	TaskProviderSync = "provider:sync"
	TaskEmailNotify  = "email:notify"
)

// ProviderSyncPayload is the payload for a provider sync task.
type ProviderSyncPayload struct {
	UserID   string `json:"userId"`
	Provider string `json:"provider"` // GITHUB, YOUTUBE
}

// NewProviderSyncTask creates a provider sync task.
func NewProviderSyncTask(userID, provider string) (*asynq.Task, error) {
	payload, err := json.Marshal(ProviderSyncPayload{UserID: userID, Provider: provider})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskProviderSync, payload, asynq.MaxRetry(3)), nil
}

// EmailNotifyPayload is the payload for an email notify task.
type EmailNotifyPayload struct {
	To       string                 `json:"to"`
	Subject  string                 `json:"subject"`
	Template string                 `json:"template"` // welcome, new_follower, job_match
	Data     map[string]interface{} `json:"data"`
}

// NewEmailNotifyTask creates an email notify task.
func NewEmailNotifyTask(to, subject, template string, data map[string]interface{}) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailNotifyPayload{To: to, Subject: subject, Template: template, Data: data})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskEmailNotify, payload, asynq.MaxRetry(5)), nil
}
