package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/provider"
	"github.com/praxis-social/praxis/server/internal/repository"
)

type AchievementService struct {
	achievementRepo *repository.AchievementRepo
	providerRepo    *repository.ProviderRepo
	userRepo        *repository.UserRepo
	skillRepo       *repository.SkillRepo
	registry        *provider.Registry
	logger          *zap.Logger
}

func NewAchievementService(
	achievementRepo *repository.AchievementRepo,
	providerRepo *repository.ProviderRepo,
	userRepo *repository.UserRepo,
	skillRepo *repository.SkillRepo,
	registry *provider.Registry,
	logger *zap.Logger,
) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		providerRepo:    providerRepo,
		userRepo:        userRepo,
		skillRepo:       skillRepo,
		registry:        registry,
		logger:          logger,
	}
}

// SyncProvider syncs achievements from a specific provider for a user.
// It returns the count of newly created achievements.
func (s *AchievementService) SyncProvider(ctx context.Context, userID, providerName string) (int, error) {
	providerName = strings.ToUpper(providerName)

	// Get the provider implementation
	p, ok := s.registry.Get(providerName)
	if !ok {
		return 0, fmt.Errorf("unknown provider: %s", providerName)
	}

	// Get the auth account for the access token
	authAccount, err := s.userRepo.GetAuthAccountByUserAndProvider(ctx, userID, strings.ToLower(providerName))
	if err != nil {
		return 0, fmt.Errorf("failed to get auth account: %w", err)
	}
	if authAccount == nil {
		return 0, fmt.Errorf("no %s account connected", providerName)
	}

	// Get user for username
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return 0, fmt.Errorf("user not found")
	}

	// Ensure connected provider record exists
	cp, err := s.providerRepo.Get(ctx, userID, providerName)
	if err != nil {
		return 0, fmt.Errorf("failed to check connected provider: %w", err)
	}
	if cp == nil {
		_, err = s.providerRepo.Create(ctx, userID, providerName, user.Username)
		if err != nil {
			return 0, fmt.Errorf("failed to create connected provider: %w", err)
		}
	}

	// Update sync status to SYNCING
	if err := s.providerRepo.UpdateSyncStatus(ctx, userID, providerName, "SYNCING"); err != nil {
		s.logger.Warn("failed to update sync status", zap.Error(err))
	}

	// Call provider.Sync
	detected, err := p.Sync(ctx, authAccount.AccessToken, user.Username)
	if err != nil {
		_ = s.providerRepo.UpdateSyncStatus(ctx, userID, providerName, "ERROR")
		return 0, fmt.Errorf("provider sync failed: %w", err)
	}

	// Archive repo-based achievements whose source no longer exists.
	// Only check persistent achievement types (repos, stars) — not ephemeral ones
	// like weekly commits or streaks, which have time-varying source IDs.
	detectedSourceIDs := make(map[string]bool, len(detected))
	for _, d := range detected {
		detectedSourceIDs[d.SourceID] = true
	}

	// Prefixes for persistent achievement types that should be archived when gone
	archivablePrefixes := []string{
		strings.ToLower(providerName) + ":repo:",
		strings.ToLower(providerName) + ":stars:",
	}

	for _, prefix := range archivablePrefixes {
		existingSourceIDs, err := s.achievementRepo.ListSourceIDsByUserAndPrefix(ctx, userID, prefix)
		if err != nil {
			s.logger.Warn("failed to list existing source IDs", zap.Error(err), zap.String("prefix", prefix))
			continue
		}

		var toArchive []string
		var toReactivate []string
		for _, sid := range existingSourceIDs {
			if detectedSourceIDs[sid] {
				toReactivate = append(toReactivate, sid)
			} else {
				toArchive = append(toArchive, sid)
			}
		}
		if err := s.achievementRepo.ArchiveBySourceIDs(ctx, toArchive); err != nil {
			s.logger.Warn("failed to archive achievements", zap.Error(err))
		} else if len(toArchive) > 0 {
			s.logger.Info("archived stale achievements", zap.Int("count", len(toArchive)), zap.String("prefix", prefix))
		}
		if err := s.achievementRepo.ReactivateBySourceIDs(ctx, toReactivate); err != nil {
			s.logger.Warn("failed to reactivate achievements", zap.Error(err))
		}
	}

	// Process detected achievements with dedup
	newCount := 0
	for _, d := range detected {
		existing, err := s.achievementRepo.GetBySourceID(ctx, d.SourceID)
		if err != nil {
			s.logger.Warn("failed to check existing achievement", zap.Error(err), zap.String("sourceId", d.SourceID))
			continue
		}
		if existing != nil {
			continue // already exists, skip
		}

		metadataJSON, err := json.Marshal(d.Metadata)
		if err != nil {
			metadataJSON = []byte("{}")
		}
		proofDataJSON, err := json.Marshal(d.ProofData)
		if err != nil {
			proofDataJSON = []byte("{}")
		}

		achievement := &repository.Achievement{
			UserID:      userID,
			Type:        d.Type,
			Title:       d.Title,
			Description: d.Description,
			Metadata:    metadataJSON,
			ProofURL:    d.ProofURL,
			ProofData:   proofDataJSON,
			Source:      providerName,
			SourceID:    d.SourceID,
			CreatedAt:   d.OccurredAt, // Use actual event time; zero value falls back to now()
		}

		_, err = s.achievementRepo.Create(ctx, achievement)
		if err != nil {
			s.logger.Warn("failed to create achievement", zap.Error(err), zap.String("sourceId", d.SourceID))
			continue
		}
		newCount++
	}

	// Extract unique languages from REPO_CREATED achievements and upsert as verified skills
	if s.skillRepo != nil {
		languages := s.extractLanguagesFromAchievements(ctx, userID)
		if len(languages) > 0 {
			var skillInputs []repository.SkillInput
			for _, lang := range languages {
				skillInputs = append(skillInputs, repository.SkillInput{
					Name:     lang,
					Verified: true,
					Source:   strings.ToLower(providerName),
				})
			}
			if err := s.skillRepo.BulkUpsert(ctx, userID, skillInputs); err != nil {
				s.logger.Warn("failed to upsert skills from provider", zap.Error(err))
			} else {
				s.logger.Info("upserted skills from provider",
					zap.String("userId", userID),
					zap.Int("count", len(skillInputs)),
				)
			}
		}
	}

	// Update sync status to IDLE
	if err := s.providerRepo.UpdateSyncStatus(ctx, userID, providerName, "IDLE"); err != nil {
		s.logger.Warn("failed to update sync status", zap.Error(err))
	}

	s.logger.Info("sync completed",
		zap.String("userId", userID),
		zap.String("provider", providerName),
		zap.Int("detected", len(detected)),
		zap.Int("new", newCount),
	)

	return newCount, nil
}

// GetFeed returns paginated feed achievements.
func (s *AchievementService) GetFeed(ctx context.Context, userID string, limit, offset int) ([]repository.AchievementWithUser, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.achievementRepo.ListFeed(ctx, userID, limit, offset)
}

// GetUserAchievements returns paginated achievements for a specific user.
func (s *AchievementService) GetUserAchievements(ctx context.Context, userID string, limit, offset int) ([]repository.AchievementWithUser, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.achievementRepo.ListByUserID(ctx, userID, limit, offset)
}

// GetAchievementByID returns a single achievement with reaction counts and user reaction.
func (s *AchievementService) GetAchievementByID(ctx context.Context, id, requestingUserID string) (map[string]interface{}, error) {
	achievement, err := s.achievementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if achievement == nil {
		return nil, nil
	}

	reactions, err := s.achievementRepo.GetReactionCounts(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get reaction counts: %w", err)
	}

	userReaction, err := s.achievementRepo.GetUserReaction(ctx, requestingUserID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reaction: %w", err)
	}

	result := map[string]interface{}{
		"achievement":  achievement,
		"reactions":    reactions,
		"userReaction": userReaction,
	}
	return result, nil
}

// GetReactionCounts returns the reaction counts for an achievement.
func (s *AchievementService) GetReactionCounts(ctx context.Context, achievementID string) (*repository.ReactionCounts, error) {
	return s.achievementRepo.GetReactionCounts(ctx, achievementID)
}

// GetUserReaction returns the user's reaction type for an achievement, or "" if none.
func (s *AchievementService) GetUserReaction(ctx context.Context, userID, achievementID string) (string, error) {
	return s.achievementRepo.GetUserReaction(ctx, userID, achievementID)
}

// RemoveReaction removes a user's reaction from an achievement.
func (s *AchievementService) RemoveReaction(ctx context.Context, userID, achievementID string) error {
	return s.achievementRepo.DeleteReaction(ctx, userID, achievementID)
}

// ToggleReaction adds or removes a reaction on an achievement.
// If the user already has a reaction, it removes it. Otherwise, it adds the new one.
func (s *AchievementService) ToggleReaction(ctx context.Context, userID, achievementID, reactionType string) error {
	existing, err := s.achievementRepo.GetUserReaction(ctx, userID, achievementID)
	if err != nil {
		return fmt.Errorf("failed to get user reaction: %w", err)
	}

	if existing == reactionType {
		// Same reaction, remove it (toggle off)
		return s.achievementRepo.DeleteReaction(ctx, userID, achievementID)
	}

	// Add or replace reaction
	return s.achievementRepo.CreateReaction(ctx, userID, achievementID, reactionType)
}

// extractLanguagesFromAchievements looks at all REPO_CREATED achievements for the user
// and extracts unique programming languages from their metadata.
func (s *AchievementService) extractLanguagesFromAchievements(ctx context.Context, userID string) []string {
	// Fetch all achievements for the user to find REPO_CREATED ones with language metadata
	achievements, err := s.achievementRepo.ListByUserID(ctx, userID, 1000, 0)
	if err != nil {
		s.logger.Warn("failed to list achievements for language extraction", zap.Error(err))
		return nil
	}

	seen := make(map[string]bool)
	var languages []string
	for _, a := range achievements {
		if a.Type != "REPO_CREATED" {
			continue
		}
		var meta map[string]interface{}
		if err := json.Unmarshal(a.Metadata, &meta); err != nil {
			continue
		}
		lang, ok := meta["language"].(string)
		if !ok || lang == "" {
			continue
		}
		if !seen[lang] {
			seen[lang] = true
			languages = append(languages, lang)
		}
	}
	return languages
}
