package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/worker"
)

// WebhookHandler handles incoming webhooks from external providers.
type WebhookHandler struct {
	asynqClient        *asynq.Client
	userRepo           *repository.UserRepo
	githubWebhookSecret string
	logger             *zap.Logger
}

// NewWebhookHandler creates a webhook handler.
func NewWebhookHandler(asynqClient *asynq.Client, userRepo *repository.UserRepo, githubWebhookSecret string, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		asynqClient:        asynqClient,
		userRepo:           userRepo,
		githubWebhookSecret: githubWebhookSecret,
		logger:             logger,
	}
}

// GitHub processes a GitHub webhook (push, repository, etc.) and enqueues a sync.
// POST /api/v1/webhooks/github
func (h *WebhookHandler) GitHub(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// Verify HMAC signature if a secret is configured.
	if h.githubWebhookSecret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if !verifyGitHubSignature(body, signature, h.githubWebhookSecret) {
			h.logger.Warn("invalid github webhook signature")
			writeError(w, http.StatusUnauthorized, "invalid signature")
			return
		}
	}

	eventType := r.Header.Get("X-GitHub-Event")
	h.logger.Info("github webhook received", zap.String("event", eventType))

	// Parse just enough to get the sender login (the GitHub user who triggered the event).
	var payload struct {
		Sender struct {
			Login string `json:"login"`
		} `json:"sender"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if payload.Sender.Login == "" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	// Find the user by GitHub username
	user, err := h.userRepo.GetUserByUsername(r.Context(), strings.ToLower(payload.Sender.Login))
	if err != nil || user == nil {
		// User isn't in our system — silently ignore.
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	task, err := worker.NewProviderSyncTask(user.ID, "GITHUB")
	if err != nil {
		h.logger.Warn("failed to create provider sync task", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if _, err := h.asynqClient.EnqueueContext(r.Context(), task, asynq.Queue("critical")); err != nil {
		h.logger.Warn("failed to enqueue github sync task", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to enqueue task")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "enqueued"})
}

// verifyGitHubSignature verifies the HMAC-SHA256 signature on a GitHub webhook payload.
// signature should be of the form "sha256=<hex>".
func verifyGitHubSignature(body []byte, signature, secret string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return false
	}
	expected := signature[len(prefix):]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	computed := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(computed), []byte(expected))
}
