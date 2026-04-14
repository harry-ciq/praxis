package provider

import (
	"context"
	"time"
)

// DetectedAchievement represents an achievement discovered by a provider during sync.
type DetectedAchievement struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProofURL    string                 `json:"proofUrl"`
	ProofData   map[string]interface{} `json:"proofData"`
	SourceID    string                 `json:"sourceId"`
	OccurredAt  time.Time              `json:"occurredAt"` // When the achievement actually happened (zero = use now())
}

// AchievementProvider defines the interface for external achievement sources.
type AchievementProvider interface {
	// ID returns the unique identifier for this provider (e.g., "GITHUB", "YOUTUBE").
	ID() string
	// Sync fetches and detects achievements from the external service.
	Sync(ctx context.Context, accessToken string, username string) ([]DetectedAchievement, error)
}
