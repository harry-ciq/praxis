package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

var starMilestones = []int{10, 50, 100, 500, 1000}

// GitHubClientFactory creates a GitHub client from an access token.
// This is exposed as a variable to allow injection in tests.
type GitHubClientFactory func(ctx context.Context, accessToken string) GitHubAPI

// GitHubAPI abstracts the GitHub client methods used by the provider.
type GitHubAPI interface {
	ListRepositories(ctx context.Context, user string, opts *github.RepositoryListByUserOptions) ([]*github.Repository, *github.Response, error)
	ListAuthenticatedUserRepos(ctx context.Context, opts *github.RepositoryListByAuthenticatedUserOptions) ([]*github.Repository, *github.Response, error)
	ListEvents(ctx context.Context, user string, opts *github.ListOptions) ([]*github.Event, *github.Response, error)
	ListCommits(ctx context.Context, owner, repo string, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
}

// defaultGitHubAPI wraps the real go-github client.
type defaultGitHubAPI struct {
	client *github.Client
}

func (d *defaultGitHubAPI) ListRepositories(ctx context.Context, user string, opts *github.RepositoryListByUserOptions) ([]*github.Repository, *github.Response, error) {
	return d.client.Repositories.ListByUser(ctx, user, opts)
}

func (d *defaultGitHubAPI) ListAuthenticatedUserRepos(ctx context.Context, opts *github.RepositoryListByAuthenticatedUserOptions) ([]*github.Repository, *github.Response, error) {
	return d.client.Repositories.ListByAuthenticatedUser(ctx, opts)
}

func (d *defaultGitHubAPI) ListCommits(ctx context.Context, owner, repo string, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	return d.client.Repositories.ListCommits(ctx, owner, repo, opts)
}

func (d *defaultGitHubAPI) ListEvents(ctx context.Context, user string, opts *github.ListOptions) ([]*github.Event, *github.Response, error) {
	return d.client.Activity.ListEventsPerformedByUser(ctx, user, false, opts)
}

// GitHubProvider detects achievements from GitHub activity.
type GitHubProvider struct {
	clientFactory GitHubClientFactory
}

// NewGitHubProvider creates a GitHub provider with the default client factory.
func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{
		clientFactory: func(ctx context.Context, accessToken string) GitHubAPI {
			client := github.NewClient(nil).WithAuthToken(accessToken)
			return &defaultGitHubAPI{client: client}
		},
	}
}

// NewGitHubProviderWithFactory creates a GitHub provider with a custom client factory (for testing).
func NewGitHubProviderWithFactory(factory GitHubClientFactory) *GitHubProvider {
	return &GitHubProvider{clientFactory: factory}
}

func (p *GitHubProvider) ID() string {
	return "GITHUB"
}

func (p *GitHubProvider) Sync(ctx context.Context, accessToken string, username string) ([]DetectedAchievement, error) {
	ghAPI := p.clientFactory(ctx, accessToken)

	var achievements []DetectedAchievement

	// 1. Fetch user's own repos (type=owner) for repo-related achievements
	repoAchievements, ownedRepos, err := p.detectRepoAchievements(ctx, ghAPI, username)
	if err != nil {
		return nil, fmt.Errorf("detecting repo achievements: %w", err)
	}
	achievements = append(achievements, repoAchievements...)

	// 2. Fetch ALL repos the user has access to (including contributed-to)
	allRepos, err := p.fetchAllRepos(ctx, ghAPI)
	if err != nil {
		// Fall back to owned repos only
		allRepos = ownedRepos
	}

	// 3. Fetch events and detect event-based achievements
	eventAchievements, err := p.detectEventAchievements(ctx, ghAPI, username)
	if err != nil {
		return nil, fmt.Errorf("detecting event achievements: %w", err)
	}
	achievements = append(achievements, eventAchievements...)

	// 4. Fetch commits per repo for weekly summary (across all repos)
	weeklyAchievements := p.detectWeeklyCommits(ctx, ghAPI, username, allRepos)
	achievements = append(achievements, weeklyAchievements...)

	return achievements, nil
}

// fetchAllRepos returns all repos the authenticated user has push access to,
// including repos they contribute to but don't own.
func (p *GitHubProvider) fetchAllRepos(ctx context.Context, api GitHubAPI) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opts := &github.RepositoryListByAuthenticatedUserOptions{
		Sort:        "pushed",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := api.ListAuthenticatedUserRepos(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("listing all repos: %w", err)
		}
		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func (p *GitHubProvider) detectRepoAchievements(ctx context.Context, api GitHubAPI, username string) ([]DetectedAchievement, []*github.Repository, error) {
	var achievements []DetectedAchievement
	var allRepos []*github.Repository

	opts := &github.RepositoryListByAuthenticatedUserOptions{
		Type:        "owner",
		Sort:        "updated",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := api.ListAuthenticatedUserRepos(ctx, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("listing repos: %w", err)
		}
		allRepos = append(allRepos, repos...)

		for _, repo := range repos {
			visibility := "public"
			if repo.GetPrivate() {
				visibility = "private"
			}

			repoCreatedAt := repo.GetCreatedAt().Time

			// REPO_CREATED achievement
			achievements = append(achievements, DetectedAchievement{
				Type:        "REPO_CREATED",
				Title:       fmt.Sprintf("Created %s", repo.GetFullName()),
				Description: truncateString(repo.GetDescription(), 200),
				Metadata: map[string]interface{}{
					"repoName":   repo.GetFullName(),
					"language":   repo.GetLanguage(),
					"stars":      repo.GetStargazersCount(),
					"visibility": visibility,
					"createdAt":  repo.GetCreatedAt().Format(time.RFC3339),
				},
				ProofURL: repo.GetHTMLURL(),
				ProofData: map[string]interface{}{
					"repoId": repo.GetID(),
				},
				SourceID:   fmt.Sprintf("github:repo:%d", repo.GetID()),
				OccurredAt: repoCreatedAt,
			})

			// STARS_MILESTONE achievements
			stars := repo.GetStargazersCount()
			for _, threshold := range starMilestones {
				if stars >= threshold {
					achievements = append(achievements, DetectedAchievement{
						Type:        "STARS_MILESTONE",
						Title:       fmt.Sprintf("%s reached %d stars", repo.GetFullName(), threshold),
						Description: fmt.Sprintf("Repository %s has reached %d stars on GitHub", repo.GetFullName(), threshold),
						Metadata: map[string]interface{}{
							"repoName":  repo.GetFullName(),
							"threshold": threshold,
							"current":   stars,
						},
						ProofURL: repo.GetHTMLURL(),
						ProofData: map[string]interface{}{
							"repoId": repo.GetID(),
							"stars":  stars,
						},
						SourceID:   fmt.Sprintf("github:stars:%s:%d", repo.GetFullName(), threshold),
						OccurredAt: repoCreatedAt,
					})
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return achievements, allRepos, nil
}

func (p *GitHubProvider) detectEventAchievements(ctx context.Context, api GitHubAPI, username string) ([]DetectedAchievement, error) {
	var achievements []DetectedAchievement

	opts := &github.ListOptions{PerPage: 100}
	var allEvents []*github.Event

	for page := 0; page < 3; page++ { // GitHub events API returns max 300 events
		events, resp, err := api.ListEvents(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("listing events: %w", err)
		}
		allEvents = append(allEvents, events...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Detect PR_MERGED events
	hasOSSContribution := false
	for _, event := range allEvents {
		if event.GetType() == "PullRequestEvent" {
			payload, err := event.ParsePayload()
			if err != nil {
				continue
			}
			prEvent, ok := payload.(*github.PullRequestEvent)
			if !ok {
				continue
			}
			pr := prEvent.GetPullRequest()
			if pr == nil || prEvent.GetAction() != "closed" || !pr.GetMerged() {
				continue
			}

			repoName := event.GetRepo().GetName()
			// Check if this is an external repo (not owned by the user)
			isExternal := event.GetRepo().GetName() != "" &&
				len(repoName) > 0 &&
				repoName[:min(len(username), len(repoName))] != username+"/"

			if isExternal {
				achievements = append(achievements, DetectedAchievement{
					Type:        "PR_MERGED",
					Title:       fmt.Sprintf("PR merged in %s", repoName),
					Description: pr.GetTitle(),
					Metadata: map[string]interface{}{
						"repoName": repoName,
						"prTitle":  pr.GetTitle(),
						"prNumber": pr.GetNumber(),
					},
					ProofURL: pr.GetHTMLURL(),
					ProofData: map[string]interface{}{
						"prId": pr.GetID(),
					},
					SourceID:   fmt.Sprintf("github:pr:%d", pr.GetID()),
					OccurredAt: pr.GetMergedAt().Time,
				})
				hasOSSContribution = true
			}
		}
	}

	// FIRST_CONTRIBUTION (first OSS PR)
	if hasOSSContribution {
		achievements = append(achievements, DetectedAchievement{
			Type:        "FIRST_CONTRIBUTION",
			Title:       "First Open Source Contribution",
			Description: fmt.Sprintf("%s made their first open source contribution", username),
			Metadata: map[string]interface{}{
				"username": username,
			},
			ProofURL:  fmt.Sprintf("https://github.com/%s", username),
			ProofData: map[string]interface{}{},
			SourceID:  fmt.Sprintf("github:first-oss:%s", username),
		})
	}

	// COMMIT_STREAK detection from PushEvents
	streakAchievements := p.detectCommitStreak(allEvents, username)
	achievements = append(achievements, streakAchievements...)

	return achievements, nil
}

func (p *GitHubProvider) detectCommitStreak(events []*github.Event, username string) []DetectedAchievement {
	// Collect unique dates with push events
	pushDates := make(map[string]bool)
	for _, event := range events {
		if event.GetType() == "PushEvent" {
			date := event.GetCreatedAt().Format("2006-01-02")
			pushDates[date] = true
		}
	}

	if len(pushDates) == 0 {
		return nil
	}

	// Sort dates
	dates := make([]string, 0, len(pushDates))
	for d := range pushDates {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	// Calculate longest streak
	maxStreak := 1
	currentStreak := 1
	for i := 1; i < len(dates); i++ {
		prev, _ := time.Parse("2006-01-02", dates[i-1])
		curr, _ := time.Parse("2006-01-02", dates[i])
		if curr.Sub(prev) == 24*time.Hour {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}

	var achievements []DetectedAchievement
	streakThresholds := []int{7, 30, 100}
	for _, threshold := range streakThresholds {
		if maxStreak >= threshold {
			achievements = append(achievements, DetectedAchievement{
				Type:        "COMMIT_STREAK",
				Title:       fmt.Sprintf("%d-day commit streak", threshold),
				Description: fmt.Sprintf("%s maintained a %d-day commit streak on GitHub", username, threshold),
				Metadata: map[string]interface{}{
					"days":      threshold,
					"maxStreak": maxStreak,
				},
				ProofURL:  fmt.Sprintf("https://github.com/%s", username),
				ProofData: map[string]interface{}{},
				SourceID:  fmt.Sprintf("github:streak:%s:%d", username, threshold),
			})
		}
	}

	return achievements
}

func (p *GitHubProvider) detectWeeklyCommits(ctx context.Context, api GitHubAPI, username string, repos []*github.Repository) []DetectedAchievement {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)

	year, week := now.ISOWeek()
	weekKey := fmt.Sprintf("%d-W%02d", year, week)

	totalCommits := 0
	type repoStat struct {
		repo    string
		commits int
		owned   bool
	}
	repoStats := make(map[string]*repoStat)

	for _, repo := range repos {
		// Only check repos pushed to in the last 7 days
		if repo.GetPushedAt().Before(weekAgo) {
			continue
		}

		owner := repo.GetOwner().GetLogin()
		repoName := repo.GetName()
		fullName := repo.GetFullName()
		isOwned := strings.EqualFold(owner, username)

		opts := &github.CommitsListOptions{
			Since:       weekAgo,
			Until:       now,
			ListOptions: github.ListOptions{PerPage: 100},
		}

		// For non-owned repos, filter by author so we only count the user's commits
		if !isOwned {
			opts.Author = username
		}

		commits, _, err := api.ListCommits(ctx, owner, repoName, opts)
		if err != nil {
			continue // skip repos we can't read commits from
		}

		count := len(commits)
		if count > 0 {
			totalCommits += count
			repoStats[fullName] = &repoStat{
				repo:    fullName,
				commits: count,
				owned:   isOwned,
			}
		}
	}

	if totalCommits == 0 {
		return nil
	}

	// Build repo breakdown for description
	topRepos := make([]map[string]interface{}, 0, len(repoStats))
	var repoNames []string
	for _, rs := range repoStats {
		topRepos = append(topRepos, map[string]interface{}{
			"repo":    rs.repo,
			"commits": rs.commits,
			"owned":   rs.owned,
		})
		repoNames = append(repoNames, rs.repo)
	}
	sort.Strings(repoNames)

	// Build a human-readable description with repo names
	desc := fmt.Sprintf("%s pushed %d commit%s across %d repo%s this week: %s",
		username, totalCommits, pluralS(totalCommits),
		len(repoStats), pluralS(len(repoStats)),
		strings.Join(repoNames, ", "))

	return []DetectedAchievement{
		{
			Type:        "WEEKLY_COMMITS",
			Title:       fmt.Sprintf("%d commits this week", totalCommits),
			Description: desc,
			Metadata: map[string]interface{}{
				"totalCommits": totalCommits,
				"repoCount":    len(repoStats),
				"repos":        topRepos,
				"weekKey":      weekKey,
			},
			ProofURL:  fmt.Sprintf("https://github.com/%s", username),
			ProofData: map[string]interface{}{},
			SourceID:  fmt.Sprintf("github:weekly-commits:%s:%s", username, weekKey),
		},
	}
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
