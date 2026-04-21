package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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

// AchievementStub is the minimum a chip / highlight needs to link back to the
// underlying achievement — inlined on the digest so the frontend doesn't need
// to make extra API calls to render a linked reveal.
type AchievementStub struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	ProofURL string `json:"proofUrl"`
	Person   string `json:"person"`   // display name of the author
	Username string `json:"username"` // author's @handle
}

// SmartFeedGroup is one themed chip in the digest.
//
// Theme is a stable tag the LLM picks from a small vocabulary so the frontend
// can color-code groups without parsing the emoji or label.
type SmartFeedGroup struct {
	Emoji        string            `json:"emoji"`
	Label        string            `json:"label"`
	Detail       string            `json:"detail"`
	Theme        string            `json:"theme"` // SHIPPING | CONTENT | MILESTONE | COMMUNITY | LEARNING | OTHER
	Achievements []AchievementStub `json:"achievements"`
}

// Milestone is a near-miss threshold that creates "just one more push" pull.
type Milestone struct {
	Emoji   string `json:"emoji"`
	Label   string `json:"label"`   // e.g. "commits to your highest week ever"
	Current int    `json:"current"` // where you are now
	Target  int    `json:"target"`  // where you need to be
}

// WatchItem is a forward-looking prediction or thing-to-watch-this-week.
type WatchItem struct {
	Emoji          string `json:"emoji"`
	Prediction     string `json:"prediction"`     // short sentence under 90 chars
	TargetUsername string `json:"targetUsername"` // optional @handle (empty = self)
}

// SmartFeedAction is a single suggested next step. When Href is set, the
// frontend renders the CTA as a link; otherwise as an advisory button.
type SmartFeedAction struct {
	Label string `json:"label"` // short description of the action
	CTA   string `json:"cta"`   // button text (e.g., "View", "React", "Follow up")
	Href  string `json:"href"`  // resolved URL (e.g., /profile/harry329). "" = no nav.
}

// FeaturedPerson is one of the people whose activity appears in this digest.
// Rendered as a small avatar chip at the top of the hero card.
type FeaturedPerson struct {
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	IsMe      bool   `json:"isMe"`
	Count     int    `json:"count"` // how many items in this digest are theirs
}

// HeroStat is the single biggest number in the digest, rendered huge with a
// count-up animation on the hero card. Chosen by the LLM from the feed.
type HeroStat struct {
	Value string `json:"value"` // Usually a number as a string, e.g. "17" or "1.2K"
	Label string `json:"label"` // Short noun phrase, e.g. "commits this week"
}

// SmartFeedDigest is the structured response the frontend consumes.
type SmartFeedDigest struct {
	// Vibe is a one-word energy tag: MOMENTUM, STEADY, EXPLORING, QUIET, MIXED.
	// Rendered as a badge above the summary.
	Vibe string `json:"vibe"`
	// HeroStat is the single biggest number in the digest (e.g. "17 commits"),
	// rendered as a large magazine-cover display element.
	HeroStat *HeroStat `json:"heroStat,omitempty"`
	// Timeframe is a short human phrase describing the window, e.g., "last 7 days".
	// Computed server-side from the oldest item in the source set.
	Timeframe string `json:"timeframe"`
	// Headline is a single attention-grabbing line, distinct from the longer summary.
	Headline string `json:"headline"`
	// Summary is the 2-3 sentence narrative in second person.
	Summary string `json:"summary"`
	// Highlight is the ONE most impressive thing this window, as a short line.
	Highlight string `json:"highlight"`
	// HighlightAchievement (if set) is the underlying record the highlight
	// refers to — used to power the "See the achievement" link.
	HighlightAchievement *AchievementStub `json:"highlightAchievement,omitempty"`
	// Groups are themed chips — 2 to 5 of them.
	Groups []SmartFeedGroup `json:"groups"`
	// SuggestedAction is one specific next step the viewer could take, or nil.
	SuggestedAction *SmartFeedAction `json:"suggestedAction,omitempty"`
	// FeaturedPeople are the unique authors in this digest, sorted by count desc.
	// Rendered as a small avatar row.
	FeaturedPeople []FeaturedPerson `json:"featuredPeople"`
	// MyShare / FollowedShare are integer percentages that sum to 100.
	// Drives the "your vs. people you follow" split bar.
	MyShare       int `json:"myShare"`
	FollowedShare int `json:"followedShare"`
	// ActivityByDay is a 7-entry array of event counts, oldest-first.
	// Drives the 7-day activity heat strip.
	ActivityByDay []int `json:"activityByDay"`
	// Milestones are near-miss thresholds — progress bars with "one more push".
	Milestones []Milestone `json:"milestones"`
	// Watching are forward-looking items worth keeping an eye on.
	Watching []WatchItem `json:"watching"`

	SourceCount int       `json:"sourceCount"`
	GeneratedAt time.Time `json:"generatedAt"`
	Cached      bool      `json:"cached"`
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
const smartFeedSystemPrompt = `You write a polished weekly digest for ONE SPECIFIC VIEWER on Praxis, a network where every achievement is verified from the real source (GitHub repos, commit streaks, PR merges, YouTube videos, subscriber/view milestones).

Each feed item is tagged with "isMine": when true the event is the viewer's own activity; when false it belongs to someone the viewer follows. Each item also has an "id" — you will reference specific items by copying their id.

Produce a JSON object with exactly these fields:

{
  "vibe": "ONE of: MOMENTUM, STEADY, EXPLORING, QUIET, MIXED. Pick based on the viewer's OWN activity intensity and variety.",
  "heroStat": {"value": "17", "label": "commits this week"},
  "headline": "A single punchy line (under 70 chars) that captures the week's energy. Written to the viewer. No emoji.",
  "summary": "2-3 sentences in second person. Use 'you' / 'your' for items where isMine is true. Name other people (first name preferred) for items where isMine is false. Be specific: counts, repo names, video titles.",
  "highlight": "Under 90 chars. The ONE most impressive thing in the feed — prefer the viewer's own if they have something noteworthy, otherwise the most interesting thing from someone they follow.",
  "highlightId": "The id of the single feed item that backs the highlight. MUST be copied exactly from one of the feed items. Empty string if highlight is empty.",
  "groups": [
    {
      "emoji": "🚀",
      "label": "Short theme",
      "detail": "one-line detail under 80 chars",
      "theme": "SHIPPING",
      "achievementIds": ["copy 1-3 exact ids from the feed items that make up this group"]
    }
  ],
  "milestones": [
    {
      "emoji": "⭐",
      "label": "under 60 chars. Phrase as 'X to <goal>'. Examples: 'stars to 100 on praxis', 'commits to your highest week ever', 'subscribers to Harry's first 10K'",
      "current": 74,
      "target": 100
    }
  ],
  "watching": [
    {
      "emoji": "👀",
      "prediction": "under 90 chars, forward-looking. Examples: 'Harry is 3 days from a week-long shipping streak', 'You're pacing 20+ commits this week'",
      "targetUsername": "optional @handle from feed items. empty string if the prediction is about the viewer themselves."
    }
  ],
  "suggestedAction": {
    "label": "ONE specific, low-friction thing the viewer could do next. Examples: 'React to Harry's new video', 'Message Sarah about her ML pipeline', 'Keep your 3-week shipping streak alive'. Under 70 chars.",
    "cta": "Short button text: React | View | Follow up | Keep going | Explore",
    "targetUsername": "The @username of the person the action refers to, copied from a feed item. MUST be exactly one of the usernames that appears in the feed items. Use empty string if the action is self-directed (about the viewer's own work)."
  }
}

Rules for heroStat:
- Pick the SINGLE most impressive number from the feed. Prefer the viewer's own activity when competitive.
- Priority: commit counts > video view/subscriber milestones > star milestones > repo counts > PR merges.
- value must be a short number string ("17", "1.2K", "50+"). Use SI abbreviations for anything >= 1000.
- label must be 2-5 words, noun phrase (e.g., "commits this week", "stars on praxis"). Start with lowercase.
- If nothing in the feed has a meaningful number, omit heroStat entirely.

Rules for groups:
- 2 to 5 groups.
- Theme must be one of: SHIPPING, CONTENT, MILESTONE, COMMUNITY, LEARNING, OTHER.
- If the viewer has their own recent activity, the FIRST group should be about them with "Your …" phrasing.
- achievementIds: copy 1-3 real ids from the feed items that drove this theme. These MUST be exact matches from the feed. Never fabricate ids.

Rules for milestones (0 to 3 items):
- Near-miss thresholds that create "one more push" motivation. Only include a milestone if you can back it with concrete numbers from the feed metadata (current counts, threshold values, streak days).
- current MUST be strictly LESS than target. If the feed shows a value already past a threshold, pick the next-higher target.
- Examples: if metadata has current=74 stars and thresholds include 100, emit {label: "stars to 100 on praxis", current: 74, target: 100}.
- If you can't find honest near-miss numbers, return an empty array. Never fabricate numbers.

Rules for watching (0 to 2 items):
- Forward-looking. Predictions MUST be grounded in current feed data (pace of commits, proximity to a known milestone threshold, days since last activity, etc.).
- Prefer predictions about the viewer themselves; include a followed user only when their data makes it interesting.
- If nothing forward-looking is honest, return an empty array.

Rules for suggestedAction:
- Pick something that creates social connection or keeps momentum. Avoid generic "share your achievements".
- Reference a specific person or piece of content from the feed when possible.

Output rules:
- Output ONLY the JSON object. No markdown fences, no commentary, no surrounding prose.
- Never invent data. Never invent ids. Never invent numbers.`

func (s *SmartFeedService) generate(ctx context.Context, viewer *repository.User, achievements []repository.AchievementWithUser) (*SmartFeedDigest, error) {
	if len(achievements) == 0 {
		return &SmartFeedDigest{
			Vibe:      "QUIET",
			Timeframe: "right now",
			Headline:  "Your feed is quiet — time to find some builders to follow.",
			Summary:   "No activity from you or the people you follow yet. Follow a few more builders to populate this digest.",
			Groups:    []SmartFeedGroup{},
			SourceCount: 0,
			GeneratedAt: time.Now(),
		}, nil
	}

	items := make([]map[string]interface{}, 0, len(achievements))
	for _, a := range achievements {
		items = append(items, map[string]interface{}{
			"id":          a.ID, // pass through to the LLM so it can reference specific items
			"person":      a.UserName,
			"username":    a.UserUsername,
			"isMine":      a.UserID == viewer.ID,
			"type":        a.Type,
			"title":       a.Title,
			"description": a.Description,
			"proofUrl":    a.ProofURL,
			"metadata":    a.Metadata, // passes through — json.RawMessage marshals verbatim
			"source":      a.Source,
			"when":        a.CreatedAt.Format(time.RFC3339),
		})
	}
	itemsJSON, _ := json.MarshalIndent(items, "", "  ")

	viewerHeader := fmt.Sprintf(
		"Viewer: %s (@%s). Write the digest to them using 'you' for items where isMine is true.\n\nFeed items (newest first):\n\n",
		viewer.Name, viewer.Username,
	)
	userText := viewerHeader + string(itemsJSON)

	system := []llm.ContentBlock{llm.CacheableBlock(smartFeedSystemPrompt)}
	userBlocks := []llm.ContentBlock{llm.TextBlock(userText)}

	raw, _, err := s.llmClient.Complete(ctx, system, userBlocks, 1200)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	cleaned := stripCodeFence(raw)

	var parsed struct {
		Vibe        string    `json:"vibe"`
		HeroStat    *HeroStat `json:"heroStat"`
		Headline    string    `json:"headline"`
		Summary     string    `json:"summary"`
		Highlight   string    `json:"highlight"`
		HighlightID string    `json:"highlightId"`
		Groups      []struct {
			Emoji          string   `json:"emoji"`
			Label          string   `json:"label"`
			Detail         string   `json:"detail"`
			Theme          string   `json:"theme"`
			AchievementIDs []string `json:"achievementIds"`
		} `json:"groups"`
		Milestones      []Milestone `json:"milestones"`
		Watching        []WatchItem `json:"watching"`
		SuggestedAction *struct {
			Label          string `json:"label"`
			CTA            string `json:"cta"`
			TargetUsername string `json:"targetUsername"`
		} `json:"suggestedAction"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		s.logger.Warn("smart feed: failed to parse LLM output as JSON",
			zap.Error(err),
			zap.String("raw", raw),
		)
		return nil, fmt.Errorf("LLM returned unparseable JSON: %w", err)
	}

	// Build an ID → achievement stub lookup so we can resolve the IDs Claude
	// copied from the feed items. Anything Claude emits that isn't in this
	// map is a hallucination and we drop it.
	stubByID := make(map[string]AchievementStub, len(achievements))
	for _, a := range achievements {
		stubByID[a.ID] = AchievementStub{
			ID:       a.ID,
			Title:    a.Title,
			Type:     a.Type,
			ProofURL: a.ProofURL,
			Person:   a.UserName,
			Username: a.UserUsername,
		}
	}

	// Transform parsed groups into the typed SmartFeedGroup the frontend sees,
	// attaching validated achievement stubs.
	groups := make([]SmartFeedGroup, 0, len(parsed.Groups))
	for _, g := range parsed.Groups {
		stubs := make([]AchievementStub, 0, len(g.AchievementIDs))
		seen := make(map[string]bool)
		for _, id := range g.AchievementIDs {
			if seen[id] {
				continue
			}
			if stub, ok := stubByID[id]; ok {
				stubs = append(stubs, stub)
				seen[id] = true
			}
			if len(stubs) == 3 {
				break
			}
		}
		groups = append(groups, SmartFeedGroup{
			Emoji:        g.Emoji,
			Label:        g.Label,
			Detail:       g.Detail,
			Theme:        normaliseTheme(g.Theme),
			Achievements: stubs,
		})
	}

	// Resolve the highlight's backing achievement (if any).
	var highlightAch *AchievementStub
	if parsed.HighlightID != "" {
		if stub, ok := stubByID[parsed.HighlightID]; ok {
			s := stub
			highlightAch = &s
		}
	}

	// Sanity-check milestones (drop invalid or impossible entries).
	milestones := make([]Milestone, 0, len(parsed.Milestones))
	for _, m := range parsed.Milestones {
		if m.Label == "" || m.Emoji == "" {
			continue
		}
		if m.Target <= 0 || m.Current < 0 || m.Current >= m.Target {
			continue
		}
		milestones = append(milestones, m)
	}

	// Sanity-check watch items.
	watching := make([]WatchItem, 0, len(parsed.Watching))
	for _, w := range parsed.Watching {
		if w.Prediction == "" {
			continue
		}
		watching = append(watching, w)
	}

	if parsed.Vibe == "" {
		parsed.Vibe = "STEADY"
	}
	parsed.Vibe = normaliseVibe(parsed.Vibe)

	// Build featured-people and counts from the raw feed (not the LLM) so we
	// can trust the data. Sort by count descending, cap at 5.
	featured, myShare, followedShare := buildFeaturedPeople(viewer, achievements)

	// Resolve the suggested action's href. Only accept usernames Claude could
	// have seen — if it hallucinated a name, drop the href.
	var action *SmartFeedAction
	if parsed.SuggestedAction != nil && parsed.SuggestedAction.Label != "" {
		action = &SmartFeedAction{
			Label: parsed.SuggestedAction.Label,
			CTA:   parsed.SuggestedAction.CTA,
		}
		if u := strings.TrimPrefix(parsed.SuggestedAction.TargetUsername, "@"); u != "" {
			validUsernames := map[string]bool{viewer.Username: true}
			for _, a := range achievements {
				validUsernames[a.UserUsername] = true
			}
			if validUsernames[u] {
				action.Href = "/profile/" + u
			}
		}
	}

	// Drop empty heroStat (Claude sometimes returns {"value":"","label":""}).
	heroStat := parsed.HeroStat
	if heroStat != nil && (heroStat.Value == "" || heroStat.Label == "") {
		heroStat = nil
	}

	return &SmartFeedDigest{
		Vibe:                 parsed.Vibe,
		HeroStat:             heroStat,
		Timeframe:            computeTimeframe(achievements),
		Headline:             parsed.Headline,
		Summary:              parsed.Summary,
		Highlight:            parsed.Highlight,
		HighlightAchievement: highlightAch,
		Groups:               groups,
		Milestones:           milestones,
		Watching:             watching,
		SuggestedAction:      action,
		FeaturedPeople:       featured,
		MyShare:              myShare,
		FollowedShare:        followedShare,
		ActivityByDay:        buildActivityByDay(achievements),
		SourceCount:          len(achievements),
		GeneratedAt:          time.Now(),
	}, nil
}

// buildFeaturedPeople returns unique authors in this digest sorted by count
// descending (max 5), plus the viewer's share of total items as an integer
// percentage. The viewer is always included first if they have any items.
func buildFeaturedPeople(viewer *repository.User, achievements []repository.AchievementWithUser) (featured []FeaturedPerson, myShare, followedShare int) {
	type bucket struct {
		Username  string
		Name      string
		AvatarURL string
		IsMe      bool
		Count     int
	}
	byUser := make(map[string]*bucket)
	myCount := 0
	for _, a := range achievements {
		b, ok := byUser[a.UserUsername]
		if !ok {
			b = &bucket{
				Username:  a.UserUsername,
				Name:      a.UserName,
				AvatarURL: a.UserAvatarURL,
				IsMe:      a.UserID == viewer.ID,
			}
			byUser[a.UserUsername] = b
		}
		b.Count++
		if b.IsMe {
			myCount++
		}
	}

	total := len(achievements)
	if total > 0 {
		myShare = (myCount * 100) / total
		followedShare = 100 - myShare
	}

	buckets := make([]*bucket, 0, len(byUser))
	for _, b := range byUser {
		buckets = append(buckets, b)
	}
	// Viewer first, then by count descending
	sort.SliceStable(buckets, func(i, j int) bool {
		if buckets[i].IsMe != buckets[j].IsMe {
			return buckets[i].IsMe
		}
		return buckets[i].Count > buckets[j].Count
	})
	if len(buckets) > 5 {
		buckets = buckets[:5]
	}

	featured = make([]FeaturedPerson, 0, len(buckets))
	for _, b := range buckets {
		featured = append(featured, FeaturedPerson{
			Username: b.Username, Name: b.Name, AvatarURL: b.AvatarURL,
			IsMe: b.IsMe, Count: b.Count,
		})
	}
	return featured, myShare, followedShare
}

// buildActivityByDay returns a 7-entry array of event counts, oldest bucket
// first (6 days ago) to newest (today). Items older than 7 days are folded
// into the first bucket so long-tail portfolios still register.
func buildActivityByDay(achievements []repository.AchievementWithUser) []int {
	out := make([]int, 7)
	today := time.Now().Truncate(24 * time.Hour)
	for _, a := range achievements {
		daysAgo := int(today.Sub(a.CreatedAt.Truncate(24 * time.Hour)).Hours() / 24)
		switch {
		case daysAgo < 0:
			daysAgo = 0
		case daysAgo > 6:
			daysAgo = 6
		}
		bucket := 6 - daysAgo // oldest first
		out[bucket]++
	}
	return out
}

// computeTimeframe returns a short phrase describing the digest window.
//
// We bias toward how *recent* the activity is rather than how far back the
// full 50-item feed stretches: a user whose 50th-most-recent achievement was
// years ago still thinks of their digest as "this week" if the top items are
// from today. We look at the newest item's age to avoid headlines like
// "last 8 years" when the feed is dominated by the last few days.
func computeTimeframe(achievements []repository.AchievementWithUser) string {
	if len(achievements) == 0 {
		return "right now"
	}
	newest := achievements[0].CreatedAt
	for _, a := range achievements {
		if a.CreatedAt.After(newest) {
			newest = a.CreatedAt
		}
	}
	hours := time.Since(newest).Hours()
	switch {
	case hours < 24:
		return "today"
	case hours < 24*7:
		return "this week"
	case hours < 24*14:
		return "the last 2 weeks"
	case hours < 24*31:
		return "this month"
	default:
		return "recent activity"
	}
}

var validVibes = map[string]bool{
	"MOMENTUM": true, "STEADY": true, "EXPLORING": true, "QUIET": true, "MIXED": true,
}

func normaliseVibe(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if validVibes[v] {
		return v
	}
	return "STEADY"
}

var validThemes = map[string]bool{
	"SHIPPING": true, "CONTENT": true, "MILESTONE": true, "COMMUNITY": true, "LEARNING": true, "OTHER": true,
}

func normaliseTheme(t string) string {
	t = strings.ToUpper(strings.TrimSpace(t))
	if validThemes[t] {
		return t
	}
	return "OTHER"
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
