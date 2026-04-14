package provider

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGitHubAPI implements GitHubAPI for testing.
type mockGitHubAPI struct {
	repos   []*github.Repository
	events  []*github.Event
	commits map[string][]*github.RepositoryCommit // key: "owner/repo"
}

func (m *mockGitHubAPI) ListRepositories(_ context.Context, _ string, _ *github.RepositoryListByUserOptions) ([]*github.Repository, *github.Response, error) {
	resp := &github.Response{NextPage: 0}
	return m.repos, resp, nil
}

func (m *mockGitHubAPI) ListAuthenticatedUserRepos(_ context.Context, _ *github.RepositoryListByAuthenticatedUserOptions) ([]*github.Repository, *github.Response, error) {
	resp := &github.Response{NextPage: 0}
	return m.repos, resp, nil
}

func (m *mockGitHubAPI) ListEvents(_ context.Context, _ string, _ *github.ListOptions) ([]*github.Event, *github.Response, error) {
	resp := &github.Response{NextPage: 0}
	return m.events, resp, nil
}

func (m *mockGitHubAPI) ListCommits(_ context.Context, owner, repo string, _ *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	key := owner + "/" + repo
	commits := m.commits[key]
	resp := &github.Response{NextPage: 0}
	return commits, resp, nil
}

func intPtr(i int) *int       { return &i }
func int64Ptr(i int64) *int64 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func newTimestamp(t time.Time) github.Timestamp {
	return github.Timestamp{Time: t}
}

func TestGitHubProvider_ID(t *testing.T) {
	p := NewGitHubProvider()
	assert.Equal(t, "GITHUB", p.ID())
}

func TestGitHubProvider_RepoCreatedAchievement(t *testing.T) {
	mock := &mockGitHubAPI{
		repos: []*github.Repository{
			{
				ID:              int64Ptr(123),
				FullName:        strPtr("testuser/myrepo"),
				Description:     strPtr("A test repo"),
				Language:        strPtr("Go"),
				StargazersCount: intPtr(5),
				HTMLURL:         strPtr("https://github.com/testuser/myrepo"),
				Private:         boolPtr(false),
				CreatedAt:       &github.Timestamp{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
		events: []*github.Event{},
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)

	// Should have exactly one REPO_CREATED achievement (no star milestones since stars=5)
	repoCreated := filterByType(achievements, "REPO_CREATED")
	assert.Len(t, repoCreated, 1)
	assert.Equal(t, "Created testuser/myrepo", repoCreated[0].Title)
	assert.Equal(t, "github:repo:123", repoCreated[0].SourceID)
	assert.Equal(t, "https://github.com/testuser/myrepo", repoCreated[0].ProofURL)
}

func TestGitHubProvider_StarMilestones(t *testing.T) {
	mock := &mockGitHubAPI{
		repos: []*github.Repository{
			{
				ID:              int64Ptr(456),
				FullName:        strPtr("testuser/popular"),
				Description:     strPtr("A popular repo"),
				Language:        strPtr("Rust"),
				StargazersCount: intPtr(120),
				HTMLURL:         strPtr("https://github.com/testuser/popular"),
				Private:         boolPtr(false),
				CreatedAt:       &github.Timestamp{Time: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
		events: []*github.Event{},
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)

	starMilestones := filterByType(achievements, "STARS_MILESTONE")
	// 120 stars should trigger milestones at 10, 50, and 100
	assert.Len(t, starMilestones, 3)

	sourceIDs := make(map[string]bool)
	for _, a := range starMilestones {
		sourceIDs[a.SourceID] = true
	}
	assert.True(t, sourceIDs["github:stars:testuser/popular:10"])
	assert.True(t, sourceIDs["github:stars:testuser/popular:50"])
	assert.True(t, sourceIDs["github:stars:testuser/popular:100"])
	assert.False(t, sourceIDs["github:stars:testuser/popular:500"])
}

func TestGitHubProvider_SourceIDFormat(t *testing.T) {
	mock := &mockGitHubAPI{
		repos: []*github.Repository{
			{
				ID:              int64Ptr(789),
				FullName:        strPtr("testuser/dedup-test"),
				Description:     strPtr("Testing dedup"),
				Language:        strPtr("Python"),
				StargazersCount: intPtr(0),
				HTMLURL:         strPtr("https://github.com/testuser/dedup-test"),
				Private:         boolPtr(false),
				CreatedAt:       &github.Timestamp{Time: time.Now()},
			},
		},
		events: []*github.Event{},
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)

	// Verify source ID format
	for _, a := range achievements {
		assert.Contains(t, a.SourceID, "github:")
	}

	repoCreated := filterByType(achievements, "REPO_CREATED")
	require.Len(t, repoCreated, 1)
	assert.Equal(t, "github:repo:789", repoCreated[0].SourceID)
}

func TestGitHubProvider_EmptyRepos(t *testing.T) {
	mock := &mockGitHubAPI{
		repos:  []*github.Repository{},
		events: []*github.Event{},
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)
	assert.Empty(t, achievements)
}

func TestGitHubProvider_PrivateReposIncluded(t *testing.T) {
	mock := &mockGitHubAPI{
		repos: []*github.Repository{
			{
				ID:              int64Ptr(100),
				FullName:        strPtr("testuser/private-repo"),
				Description:     strPtr("Secret stuff"),
				Language:        strPtr("Go"),
				StargazersCount: intPtr(0),
				HTMLURL:         strPtr("https://github.com/testuser/private-repo"),
				Private:         boolPtr(true),
				CreatedAt:       &github.Timestamp{Time: time.Now()},
			},
		},
		events: []*github.Event{},
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)
	require.Len(t, achievements, 1)
	assert.Equal(t, "REPO_CREATED", achievements[0].Type)
	assert.Equal(t, "private", achievements[0].Metadata["visibility"])
}

func TestGitHubProvider_CommitStreak(t *testing.T) {
	// Create push events for 8 consecutive days
	now := time.Now()
	var events []*github.Event
	for i := 0; i < 8; i++ {
		eventTime := now.AddDate(0, 0, -i)
		ts := newTimestamp(eventTime)
		events = append(events, &github.Event{
			Type:      strPtr("PushEvent"),
			CreatedAt: &ts,
		})
	}

	mock := &mockGitHubAPI{
		repos:  []*github.Repository{},
		events: events,
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)

	streaks := filterByType(achievements, "COMMIT_STREAK")
	// 8-day streak should trigger 7-day threshold only
	assert.Len(t, streaks, 1)
	assert.Equal(t, "github:streak:testuser:7", streaks[0].SourceID)
}

func TestGitHubProvider_WeeklyCommitsContributedRepo(t *testing.T) {
	now := time.Now()
	pushedAt := github.Timestamp{Time: now}

	mock := &mockGitHubAPIWithAuthorCapture{
		mockGitHubAPI: mockGitHubAPI{
			repos: []*github.Repository{
				{
					ID:              int64Ptr(300),
					FullName:        strPtr("otheruser/cool-project"),
					Name:            strPtr("cool-project"),
					Description:     strPtr("Someone else's project"),
					Language:        strPtr("TypeScript"),
					StargazersCount: intPtr(50),
					HTMLURL:         strPtr("https://github.com/otheruser/cool-project"),
					Private:         boolPtr(false),
					CreatedAt:       &github.Timestamp{Time: now.AddDate(0, -6, 0)},
					PushedAt:        &pushedAt,
					Owner:           &github.User{Login: strPtr("otheruser")},
				},
			},
			events: []*github.Event{},
			commits: map[string][]*github.RepositoryCommit{
				"otheruser/cool-project": {
					{SHA: strPtr("aaa111")},
					{SHA: strPtr("bbb222")},
				},
			},
		},
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)

	weekly := filterByType(achievements, "WEEKLY_COMMITS")
	require.Len(t, weekly, 1)
	assert.Equal(t, "2 commits this week", weekly[0].Title)
	assert.Contains(t, weekly[0].Description, "otheruser/cool-project")

	// Verify the author filter was set for non-owned repos
	assert.Equal(t, "testuser", mock.lastCommitAuthor, "should filter by author for non-owned repos")

	// Check metadata has owned=false
	repos := weekly[0].Metadata["repos"].([]map[string]interface{})
	require.Len(t, repos, 1)
	assert.Equal(t, false, repos[0]["owned"])
}

// mockGitHubAPIWithAuthorCapture extends mockGitHubAPI to capture commit list options.
type mockGitHubAPIWithAuthorCapture struct {
	mockGitHubAPI
	lastCommitAuthor string
}

func (m *mockGitHubAPIWithAuthorCapture) ListCommits(_ context.Context, owner, repo string, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	m.lastCommitAuthor = opts.Author
	key := owner + "/" + repo
	commits := m.commits[key]
	resp := &github.Response{NextPage: 0}
	return commits, resp, nil
}

func TestGitHubProvider_WeeklyCommits(t *testing.T) {
	now := time.Now()
	pushedAt := github.Timestamp{Time: now}

	mock := &mockGitHubAPI{
		repos: []*github.Repository{
			{
				ID:              int64Ptr(200),
				FullName:        strPtr("testuser/active-repo"),
				Name:            strPtr("active-repo"),
				Description:     strPtr("Active project"),
				Language:        strPtr("Go"),
				StargazersCount: intPtr(0),
				HTMLURL:         strPtr("https://github.com/testuser/active-repo"),
				Private:         boolPtr(false),
				CreatedAt:       &github.Timestamp{Time: now.AddDate(0, -1, 0)},
				PushedAt:        &pushedAt,
				Owner:           &github.User{Login: strPtr("testuser")},
			},
		},
		events: []*github.Event{},
		commits: map[string][]*github.RepositoryCommit{
			"testuser/active-repo": {
				{SHA: strPtr("abc123")},
				{SHA: strPtr("def456")},
				{SHA: strPtr("ghi789")},
			},
		},
	}

	p := NewGitHubProviderWithFactory(func(_ context.Context, _ string) GitHubAPI {
		return mock
	})

	achievements, err := p.Sync(context.Background(), "token", "testuser")
	require.NoError(t, err)

	weekly := filterByType(achievements, "WEEKLY_COMMITS")
	require.Len(t, weekly, 1)
	assert.Equal(t, "3 commits this week", weekly[0].Title)
	assert.Contains(t, weekly[0].SourceID, "github:weekly-commits:testuser:")
}

func filterByType(achievements []DetectedAchievement, achievementType string) []DetectedAchievement {
	var result []DetectedAchievement
	for _, a := range achievements {
		if a.Type == achievementType {
			result = append(result, a)
		}
	}
	return result
}
