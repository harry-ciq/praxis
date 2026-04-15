package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID          string          `json:"id"`
	Username    string          `json:"username"`
	Email       string          `json:"email"`
	Name        string          `json:"name"`
	Bio         string          `json:"bio"`
	AvatarURL   string          `json:"avatarUrl"`
	Headline    string          `json:"headline"`
	Location    string          `json:"location"`
	WebsiteURL  string          `json:"websiteUrl"`
	SocialLinks json.RawMessage `json:"socialLinks"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type AuthAccount struct {
	ID                string     `json:"id"`
	UserID            string     `json:"userId"`
	Provider          string     `json:"provider"`
	ProviderAccountID string     `json:"providerAccountId"`
	AccessToken       string     `json:"-"`
	RefreshToken      string     `json:"-"`
	ExpiresAt         *time.Time `json:"expiresAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) CreateUser(ctx context.Context, username, email, name, avatarURL string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, name, avatar_url)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, email, name, bio, avatar_url, headline, location, website_url, social_links, created_at, updated_at`,
		username, email, name, avatarURL,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Bio, &u.AvatarURL, &u.Headline, &u.Location, &u.WebsiteURL, &u.SocialLinks, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, name, bio, avatar_url, headline, location, website_url, social_links, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Bio, &u.AvatarURL, &u.Headline, &u.Location, &u.WebsiteURL, &u.SocialLinks, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, name, bio, avatar_url, headline, location, website_url, social_links, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Bio, &u.AvatarURL, &u.Headline, &u.Location, &u.WebsiteURL, &u.SocialLinks, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, name, bio, avatar_url, headline, location, website_url, social_links, created_at, updated_at
		 FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Bio, &u.AvatarURL, &u.Headline, &u.Location, &u.WebsiteURL, &u.SocialLinks, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, id string, name, bio, headline, location, websiteURL *string, socialLinks *json.RawMessage) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`UPDATE users
		 SET name = COALESCE($2, name),
		     bio = COALESCE($3, bio),
		     headline = COALESCE($4, headline),
		     location = COALESCE($5, location),
		     website_url = COALESCE($6, website_url),
		     social_links = COALESCE($7, social_links),
		     updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, username, email, name, bio, avatar_url, headline, location, website_url, social_links, created_at, updated_at`,
		id, name, bio, headline, location, websiteURL, socialLinks,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Bio, &u.AvatarURL, &u.Headline, &u.Location, &u.WebsiteURL, &u.SocialLinks, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) CreateAuthAccount(ctx context.Context, userID, provider, providerAccountID, accessToken, refreshToken string) (*AuthAccount, error) {
	var a AuthAccount
	err := r.pool.QueryRow(ctx,
		`INSERT INTO auth_accounts (user_id, provider, provider_account_id, access_token, refresh_token)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, provider, provider_account_id, access_token, refresh_token, expires_at, created_at`,
		userID, provider, providerAccountID, accessToken, refreshToken,
	).Scan(&a.ID, &a.UserID, &a.Provider, &a.ProviderAccountID, &a.AccessToken, &a.RefreshToken, &a.ExpiresAt, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *UserRepo) GetAuthAccount(ctx context.Context, provider, providerAccountID string) (*AuthAccount, error) {
	var a AuthAccount
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, provider_account_id, access_token, refresh_token, expires_at, created_at
		 FROM auth_accounts WHERE provider = $1 AND provider_account_id = $2`,
		provider, providerAccountID,
	).Scan(&a.ID, &a.UserID, &a.Provider, &a.ProviderAccountID, &a.AccessToken, &a.RefreshToken, &a.ExpiresAt, &a.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *UserRepo) GetAuthAccountByUserAndProvider(ctx context.Context, userID, provider string) (*AuthAccount, error) {
	var a AuthAccount
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, provider_account_id, access_token, refresh_token, expires_at, created_at
		 FROM auth_accounts WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	).Scan(&a.ID, &a.UserID, &a.Provider, &a.ProviderAccountID, &a.AccessToken, &a.RefreshToken, &a.ExpiresAt, &a.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *UserRepo) UpdateAuthTokens(ctx context.Context, id, accessToken, refreshToken string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE auth_accounts SET access_token = $1, refresh_token = $2 WHERE id = $3`,
		accessToken, refreshToken, id,
	)
	return err
}
