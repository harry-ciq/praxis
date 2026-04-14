package service

import (
	"context"
	"errors"

	"github.com/praxis-social/praxis/server/internal/repository"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)

type UserService struct {
	userRepo        *repository.UserRepo
	followRepo      *repository.FollowRepo
	achievementRepo *repository.AchievementRepo
	providerRepo    *repository.ProviderRepo
}

type UserProfileResponse struct {
	User             *repository.User               `json:"user"`
	AchievementCount int                            `json:"achievementCount"`
	FollowerCount    int                            `json:"followerCount"`
	FollowingCount   int                            `json:"followingCount"`
	IsFollowing      bool                           `json:"isFollowing"`
	Providers        []repository.ConnectedProvider `json:"providers"`
}

type UpdateProfileRequest struct {
	Name     *string `json:"name"`
	Bio      *string `json:"bio"`
	Headline *string `json:"headline"`
}

func NewUserService(
	userRepo *repository.UserRepo,
	followRepo *repository.FollowRepo,
	achievementRepo *repository.AchievementRepo,
	providerRepo *repository.ProviderRepo,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		followRepo:      followRepo,
		achievementRepo: achievementRepo,
		providerRepo:    providerRepo,
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

	var isFollowing bool
	if viewerID != "" && viewerID != user.ID {
		isFollowing, err = s.followRepo.IsFollowing(ctx, viewerID, user.ID)
		if err != nil {
			return nil, err
		}
	}

	return &UserProfileResponse{
		User:             user,
		AchievementCount: achievementCount,
		FollowerCount:    followers,
		FollowingCount:   following,
		IsFollowing:      isFollowing,
		Providers:        providers,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*repository.User, error) {
	return s.userRepo.UpdateUser(ctx, userID, req.Name, req.Bio, req.Headline)
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
