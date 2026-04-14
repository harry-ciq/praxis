package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"github.com/praxis-social/praxis/server/internal/config"
	"github.com/praxis-social/praxis/server/internal/repository"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
	redisRefreshPrefix   = "refresh:"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	redis     *redis.Client
	config    *config.Config
	jwtSecret []byte
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type gitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type gitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func NewAuthService(userRepo *repository.UserRepo, rdb *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		redis:     rdb,
		config:    cfg,
		jwtSecret: []byte(cfg.JWTSecret),
	}
}

// GenerateTokenPair creates access (15min) and refresh (7day) JWT tokens.
func (s *AuthService) GenerateTokenPair(user *repository.User) (*TokenPair, error) {
	now := time.Now()

	// Access token
	accessClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store refresh token in Redis
	err = s.redis.Set(context.Background(), redisRefreshPrefix+refreshToken, user.ID, refreshTokenDuration).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken parses and validates an access token, returning its claims.
func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// RefreshTokens validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Verify the refresh token exists in Redis
	userID, err := s.redis.Get(ctx, redisRefreshPrefix+refreshToken).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check refresh token: %w", err)
	}

	// Validate the JWT itself
	claims, err := s.ValidateAccessToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.UserID != userID {
		return nil, fmt.Errorf("token user mismatch")
	}

	// Get the user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Invalidate old refresh token
	s.redis.Del(ctx, redisRefreshPrefix+refreshToken)

	// Generate new pair
	return s.GenerateTokenPair(user)
}

// HandleGitHubCallback exchanges a code for tokens, fetches the GitHub user,
// creates or finds the local user, and returns a JWT token pair.
func (s *AuthService) HandleGitHubCallback(ctx context.Context, code string) (*TokenPair, *repository.User, error) {
	// 1. Exchange code for GitHub access token
	ghToken, err := s.exchangeGitHubCode(code)
	if err != nil {
		return nil, nil, fmt.Errorf("github code exchange failed: %w", err)
	}

	// 2. Fetch GitHub user info
	ghUser, err := s.getGitHubUser(ghToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get github user: %w", err)
	}

	ghAccountID := strconv.Itoa(ghUser.ID)

	// 3. Check if auth_account exists
	authAccount, err := s.userRepo.GetAuthAccount(ctx, "github", ghAccountID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check auth account: %w", err)
	}

	var user *repository.User

	if authAccount != nil {
		// 4a. Existing account - get user and update tokens
		user, err = s.userRepo.GetUserByID(ctx, authAccount.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get user: %w", err)
		}
		if user == nil {
			return nil, nil, fmt.Errorf("user not found for auth account")
		}

		err = s.userRepo.UpdateAuthTokens(ctx, authAccount.ID, ghToken, "")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update auth tokens: %w", err)
		}
	} else {
		// 4b. New account - create user and auth_account
		email := ghUser.Email
		if email == "" {
			email = ghUser.Login + "@github.com"
		}
		name := ghUser.Name
		if name == "" {
			name = ghUser.Login
		}

		user, err = s.userRepo.CreateUser(ctx, ghUser.Login, email, name, ghUser.AvatarURL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create user: %w", err)
		}

		_, err = s.userRepo.CreateAuthAccount(ctx, user.ID, "github", ghAccountID, ghToken, "")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create auth account: %w", err)
		}
	}

	// 5. Generate JWT token pair
	tokenPair, err := s.GenerateTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokenPair, user, nil
}

// Logout invalidates the refresh token in Redis.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.redis.Del(ctx, redisRefreshPrefix+refreshToken).Err()
}

// GitHubAuthURL returns the GitHub OAuth authorization URL.
func (s *AuthService) GitHubAuthURL() string {
	params := url.Values{
		"client_id":    {s.config.GitHubClientID},
		"redirect_uri": {s.config.FrontendURL + "/auth/callback"},
		"scope":        {"read:user user:email repo"},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (s *AuthService) exchangeGitHubCode(code string) (string, error) {
	data := url.Values{
		"client_id":     {s.config.GitHubClientID},
		"client_secret": {s.config.GitHubClientSecret},
		"code":          {code},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp gitHubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token from github")
	}

	return tokenResp.AccessToken, nil
}

func (s *AuthService) getGitHubUser(accessToken string) (*gitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get github user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned %d: %s", resp.StatusCode, string(body))
	}

	var ghUser gitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, fmt.Errorf("failed to decode github user: %w", err)
	}

	return &ghUser, nil
}
