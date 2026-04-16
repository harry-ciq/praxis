package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/praxis-social/praxis/server/internal/repository"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)

type UserService struct {
	userRepo        *repository.UserRepo
	followRepo      *repository.FollowRepo
	achievementRepo *repository.AchievementRepo
	providerRepo    *repository.ProviderRepo
	experienceRepo  *repository.ExperienceRepo
	skillRepo       *repository.SkillRepo
}

type UserProfileResponse struct {
	ID               string                         `json:"id"`
	Username         string                         `json:"username"`
	Email            string                         `json:"email"`
	Name             string                         `json:"name"`
	Bio              string                         `json:"bio"`
	AvatarURL        string                         `json:"avatarUrl"`
	Headline         string                         `json:"headline"`
	Location         string                         `json:"location"`
	WebsiteURL       string                         `json:"websiteUrl"`
	SocialLinks      json.RawMessage                `json:"socialLinks"`
	CreatedAt        string                         `json:"createdAt"`
	AchievementCount int                            `json:"achievementCount"`
	FollowerCount    int                            `json:"followerCount"`
	FollowingCount   int                            `json:"followingCount"`
	IsFollowing      bool                           `json:"isFollowing"`
	Providers        []repository.ConnectedProvider `json:"providers"`
	Experiences      []repository.Experience        `json:"experiences"`
	Skills           []repository.Skill             `json:"skills"`
}

type UserSummary struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatarUrl"`
	Headline    string `json:"headline"`
	IsFollowing bool   `json:"isFollowing"`
}

type UpdateProfileRequest struct {
	Name        *string          `json:"name"`
	Bio         *string          `json:"bio"`
	Headline    *string          `json:"headline"`
	Location    *string          `json:"location"`
	WebsiteURL  *string          `json:"websiteUrl"`
	SocialLinks *json.RawMessage `json:"socialLinks"`
}

func NewUserService(
	userRepo *repository.UserRepo,
	followRepo *repository.FollowRepo,
	achievementRepo *repository.AchievementRepo,
	providerRepo *repository.ProviderRepo,
	experienceRepo *repository.ExperienceRepo,
	skillRepo *repository.SkillRepo,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		followRepo:      followRepo,
		achievementRepo: achievementRepo,
		providerRepo:    providerRepo,
		experienceRepo:  experienceRepo,
		skillRepo:       skillRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, username string, viewerID string) (*UserProfileResponse, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	achievementCount, err := s.achievementRepo.CountByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	followers, following, err := s.followRepo.GetCounts(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	providers, err := s.providerRepo.List(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	experiences, err := s.experienceRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	skills, err := s.skillRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	var isFollowing bool
	if viewerID != "" && viewerID != user.ID {
		isFollowing, err = s.followRepo.IsFollowing(ctx, viewerID, user.ID)
		if err != nil {
			return nil, err
		}
	}

	socialLinks := user.SocialLinks
	if socialLinks == nil {
		socialLinks = json.RawMessage("{}")
	}

	return &UserProfileResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		Name:             user.Name,
		Bio:              user.Bio,
		AvatarURL:        user.AvatarURL,
		Headline:         user.Headline,
		Location:         user.Location,
		WebsiteURL:       user.WebsiteURL,
		SocialLinks:      socialLinks,
		CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		AchievementCount: achievementCount,
		FollowerCount:    followers,
		FollowingCount:   following,
		IsFollowing:      isFollowing,
		Providers:        providers,
		Experiences:      experiences,
		Skills:           skills,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*repository.User, error) {
	return s.userRepo.UpdateUser(ctx, userID, req.Name, req.Bio, req.Headline, req.Location, req.WebsiteURL, req.SocialLinks)
}

func (s *UserService) Follow(ctx context.Context, followerID, username string) error {
	target, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	if followerID == target.ID {
		return ErrCannotFollowSelf
	}
	return s.followRepo.Follow(ctx, followerID, target.ID)
}

func (s *UserService) Unfollow(ctx context.Context, followerID, username string) error {
	target, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	return s.followRepo.Unfollow(ctx, followerID, target.ID)
}

func (s *UserService) GetFollowers(ctx context.Context, username string, viewerID string, limit, offset int) ([]UserSummary, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	users, err := s.followRepo.ListFollowers(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.toUserSummaries(ctx, users, viewerID)
}

func (s *UserService) GetFollowing(ctx context.Context, username string, viewerID string, limit, offset int) ([]UserSummary, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	users, err := s.followRepo.ListFollowing(ctx, user.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.toUserSummaries(ctx, users, viewerID)
}

func (s *UserService) SearchUsers(ctx context.Context, query string, viewerID string, limit, offset int) ([]UserSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	users, err := s.userRepo.SearchUsers(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.toUserSummaries(ctx, users, viewerID)
}

func (s *UserService) toUserSummaries(ctx context.Context, users []repository.User, viewerID string) ([]UserSummary, error) {
	summaries := make([]UserSummary, 0, len(users))
	for _, u := range users {
		var isFollowing bool
		if viewerID != "" && viewerID != u.ID {
			var err error
			isFollowing, err = s.followRepo.IsFollowing(ctx, viewerID, u.ID)
			if err != nil {
				return nil, err
			}
		}
		summaries = append(summaries, UserSummary{
			ID:          u.ID,
			Username:    u.Username,
			Name:        u.Name,
			AvatarURL:   u.AvatarURL,
			Headline:    u.Headline,
			IsFollowing: isFollowing,
		})
	}
	return summaries, nil
}
