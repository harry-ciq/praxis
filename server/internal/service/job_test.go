package service

import (
	"encoding/json"
	"testing"

	"github.com/praxis-social/praxis/server/internal/repository"
)

func ach(t string, meta map[string]interface{}) repository.AchievementWithUser {
	b, _ := json.Marshal(meta)
	return repository.AchievementWithUser{
		Achievement: repository.Achievement{
			Type:     t,
			Metadata: b,
		},
	}
}

func TestScoreRequirements(t *testing.T) {
	user := []repository.AchievementWithUser{
		ach("REPO_CREATED", map[string]interface{}{"repoName": "a"}),
		ach("REPO_CREATED", map[string]interface{}{"repoName": "b"}),
		ach("REPO_CREATED", map[string]interface{}{"repoName": "c"}),
		ach("STARS_MILESTONE", map[string]interface{}{"current": 150, "threshold": 100}),
		ach("COMMIT_STREAK", map[string]interface{}{"days": 7, "maxStreak": 12}),
		ach("VIDEO_PUBLISHED", map[string]interface{}{}),
	}

	cases := []struct {
		name        string
		reqs        []string
		wantMatched int
	}{
		{"empty", []string{}, 0},
		{"type only - present", []string{"REPO_CREATED"}, 1},
		{"type only - absent", []string{"PR_MERGED"}, 0},
		{"count threshold - met", []string{"REPO_CREATED:3"}, 1},
		{"count threshold - exceeded ok", []string{"REPO_CREATED:2"}, 1},
		{"count threshold - missed", []string{"REPO_CREATED:5"}, 0},
		{"stars - met by current", []string{"STARS_MILESTONE:100"}, 1},
		{"stars - met by current with lower bar", []string{"STARS_MILESTONE:50"}, 1},
		{"stars - missed", []string{"STARS_MILESTONE:200"}, 0},
		{"streak - met by maxStreak", []string{"COMMIT_STREAK:10"}, 1},
		{"streak - missed", []string{"COMMIT_STREAK:30"}, 0},
		{"mix", []string{"REPO_CREATED", "STARS_MILESTONE:100", "PR_MERGED:5"}, 2},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _ := scoreRequirements(c.reqs, user)
			if got != c.wantMatched {
				t.Errorf("scoreRequirements: want %d, got %d", c.wantMatched, got)
			}
		})
	}
}

func TestScoreRequirements_MatchedListIncludesOriginalString(t *testing.T) {
	user := []repository.AchievementWithUser{
		ach("STARS_MILESTONE", map[string]interface{}{"current": 200}),
	}
	_, list := scoreRequirements([]string{"STARS_MILESTONE:50"}, user)
	if len(list) != 1 || list[0] != "STARS_MILESTONE:50" {
		t.Errorf("expected matched list to include original requirement string, got %v", list)
	}
}
