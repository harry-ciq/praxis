package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/llm"
	"github.com/praxis-social/praxis/server/internal/repository"
)

// ErrSmartFeedUnavailable is returned when the LLM client isn't configured
// (e.g. no ANTHROPIC_API_KEY in this environment).
var ErrSmartFeedUnavailable = errors.New("smart feed unavailable")

// SmartFeedGroup is one themed chip in the digest.
type SmartFeedGroup struct {
	Emoji  string `json:"emoji"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
}

// SmartFeedDigest is the structured response the frontend consumes.
type SmartFeedDigest struct {
	Summary     string           `json:"summary"`
	Groups      []SmartFeedGroup `json:"groups"`
	SourceCount int              `json:"sourceCount"`
	GeneratedAt time.Time        `json:"generatedAt"`
	Cached      bool             `json:"cached"`
}

// SmartFeedService computes an LLM-assisted digest of a user's feed.
type SmartFeedService struct {
	achievementRepo *repository.AchievementRepo
	userRepo        *repository.UserRepo
	llmClient       *llm.Client
	redis           *redis.Client
	cacheTTL        time.Duration
	logger          *zap.Logger
}

func NewSmartFeedService(
	achievementRepo *repository.AchievementRepo,
	userRepo *repository.UserRepo,
	llmClient *llm.Client,
	rdb *redis.Client,
	cacheTTLSeconds int,
	logger *zap.Logger,
) *SmartFeedService {
	ttl := time.Duration(cacheTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &SmartFeedService{
		achievementRepo: achievementRepo,
		userRepo:        userRepo,
		llmClient:       llmClient,
		redis:           rdb,
		cacheTTL:        ttl,
		logger:          logger,
	}
}

// Generate returns a smart-feed digest personalized for the given viewer.
// Same feed scope as the raw feed (self + followed users), but the prompt is
// rewritten from the viewer's POV so two users with overlapping feeds each get
// their own first-person narrative. Redis-cached per viewer with configurable
// TTL. If force is true, bypasses the cache.
func (s *SmartFeedService) Generate(ctx context.Context, userID string, force bool) (*SmartFeedDigest, error) {
	if !s.llmClient.IsConfigured() {
		return nil, ErrSmartFeedUnavailable
	}

	cacheKey := "smart-feed:" + userID

	if !force {
		if cached, err := s.readCache(ctx, cacheKey); err == nil && cached != nil {
			cached.Cached = true
			return cached, nil
		}
	}

	viewer, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load viewer: %w", err)
	}
	if viewer == nil {
		return nil, fmt.Errorf("viewer not found: %s", userID)
	}

	achievements, err := s.achievementRepo.ListFeed(ctx, userID, 50, 0)
	if err != nil {
		return nil, fmt.Errorf("list feed: %w", err)
	}

	digest, err := s.generate(ctx, viewer, achievements)
	if err != nil {
		return nil, err
	}

	_ = s.writeCache(ctx, cacheKey, digest)
	return digest, nil
}

// InvalidateCache drops the cached digest for a viewer. Call this after events
// that change what a viewer's feed should look like (follow/unfollow, new sync).
func (s *SmartFeedService) InvalidateCache(ctx context.Context, userID string) {
	_ = s.redis.Del(ctx, "smart-feed:"+userID).Err()
}

func (s *SmartFeedService) readCache(ctx context.Context, key string) (*SmartFeedDigest, error) {
	raw, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var d SmartFeedDigest
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *SmartFeedService) writeCache(ctx context.Context, key string, d *SmartFeedDigest) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, key, b, s.cacheTTL).Err()
}

// smartFeedSystemPrompt is the instruction to Claude. The rules are stable
// across users so this block is marked as cacheable (5-min ephemeral). The
// viewer's identity is injected in the user message so it doesn't bust the
// system-prompt cache.
const smartFeedSystemPrompt = `You summarize a social feed of verified professional achievements on Praxis, a network where every event is real (repo creation, star milestones, YouTube videos published, subscriber milestones, commit streaks, PR merges).

You write the summary FOR ONE SPECIFIC VIEWER. Each feed item is tagged with "isMine": when true the event is the viewer's own activity; when false it belongs to someone the viewer follows.

Produce a JSON object with exactly these fields:

{
  "summary": "2-3 sentences written TO the viewer in second person. Use 'you' / 'your' for items where isMine is true. Use the other person's name (first name preferred) for items where isMine is false. Be specific: mention counts, repo names, video titles. No marketing fluff.",
  "groups": [
    {"emoji": "🚀", "label": "Short theme", "detail": "one-line detail. Prefer separating 'Your …' from activity by people you follow when both exist."}
  ]
}

Rules:
- Output ONLY the JSON object. No markdown fences, no commentary.
- 2 to 5 groups. Each emoji/label/detail must be short (detail < 80 chars).
- If the viewer has their own recent activity, the first group should reflect that with "Your ..." phrasing.
- Good themes: Shipping (repos/commits), Content (videos/articles), Milestones (stars/subscribers/views), Community (PRs merged, first OSS).
- If the feed is empty or trivial, return a short summary saying so and an empty groups array.
- Never invent data. Only use facts that appear in the feed items.`

func (s *SmartFeedService) generate(ctx context.Context, viewer *repository.User, achievements []repository.AchievementWithUser) (*SmartFeedDigest, error) {
	if len(achievements) == 0 {
		return &SmartFeedDigest{
			Summary:     "Your feed is quiet right now. Follow some builders to see their achievements here.",
			Groups:      []SmartFeedGroup{},
			SourceCount: 0,
			GeneratedAt: time.Now(),
		}, nil
	}

	items := make([]map[string]interface{}, 0, len(achievements))
	for _, a := range achievements {
		items = append(items, map[string]interface{}{
			"person":      a.UserName,
			"username":    a.UserUsername,
			"isMine":      a.UserID == viewer.ID,
			"type":        a.Type,
			"title":       a.Title,
			"description": a.Description,
			"source":      a.Source,
			"when":        a.CreatedAt.Format(time.RFC3339),
		})
	}
	itemsJSON, _ := json.MarshalIndent(items, "", "  ")

	viewerHeader := fmt.Sprintf(
		"Viewer: %s (@%s). Write the summary to them using 'you' for items where isMine is true.\n\nFeed items (newest first):\n\n",
		viewer.Name, viewer.Username,
	)
	userText := viewerHeader + string(itemsJSON)

	system := []llm.ContentBlock{llm.CacheableBlock(smartFeedSystemPrompt)}
	userBlocks := []llm.ContentBlock{llm.TextBlock(userText)}

	raw, _, err := s.llmClient.Complete(ctx, system, userBlocks, 800)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	// Claude sometimes wraps in ```json ... ``` even when instructed not to.
	cleaned := stripCodeFence(raw)

	var parsed struct {
		Summary string           `json:"summary"`
		Groups  []SmartFeedGroup `json:"groups"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		s.logger.Warn("smart feed: failed to parse LLM output as JSON",
			zap.Error(err),
			zap.String("raw", raw),
		)
		return nil, fmt.Errorf("LLM returned unparseable JSON: %w", err)
	}

	if parsed.Groups == nil {
		parsed.Groups = []SmartFeedGroup{}
	}

	return &SmartFeedDigest{
		Summary:     parsed.Summary,
		Groups:      parsed.Groups,
		SourceCount: len(achievements),
		GeneratedAt: time.Now(),
	}, nil
}

// stripCodeFence removes leading/trailing ```json fences and whitespace.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// drop the first fence line
		if idx := strings.Index(s, "\n"); idx >= 0 {
			s = s[idx+1:]
		}
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
