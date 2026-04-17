package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var subscriberMilestones = []int{10, 50, 100, 500, 1000, 5000, 10000, 50000, 100000, 1000000}
var viewMilestones = []int{1000, 10000, 100000, 1000000, 10000000}

// YouTubeAPI abstracts the YouTube Data API methods used by the provider.
type YouTubeAPI interface {
	GetChannel(ctx context.Context) (*ytChannel, error)
	ListVideos(ctx context.Context, channelID string, maxResults int) ([]ytVideo, error)
}

type ytChannel struct {
	ID              string
	Title           string
	Description     string
	CustomURL       string
	PublishedAt     time.Time
	SubscriberCount int
	VideoCount      int
	ViewCount       int
	ThumbnailURL    string
}

type ytVideo struct {
	ID          string
	Title       string
	Description string
	PublishedAt time.Time
	ViewCount   int
	LikeCount   int
	CommentCount int
	ThumbnailURL string
	ChannelTitle string
}

// YouTubeClientFactory creates a YouTube API client from an access token.
type YouTubeClientFactory func(ctx context.Context, accessToken string) YouTubeAPI

// defaultYouTubeAPI implements YouTubeAPI using the YouTube Data API v3 REST endpoints.
type defaultYouTubeAPI struct {
	accessToken string
	httpClient  *http.Client
}

func (d *defaultYouTubeAPI) doRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	base := "https://www.googleapis.com/youtube/v3"
	u := fmt.Sprintf("%s/%s?%s", base, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube API returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (d *defaultYouTubeAPI) GetChannel(ctx context.Context) (*ytChannel, error) {
	params := url.Values{
		"part": {"snippet,statistics"},
		"mine": {"true"},
	}

	body, err := d.doRequest(ctx, "channels", params)
	if err != nil {
		return nil, fmt.Errorf("fetching channel: %w", err)
	}

	var resp struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				CustomURL   string `json:"customUrl"`
				PublishedAt string `json:"publishedAt"`
				Thumbnails  struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount       string `json:"viewCount"`
				SubscriberCount string `json:"subscriberCount"`
				VideoCount      string `json:"videoCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing channel response: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no YouTube channel found for this account")
	}

	item := resp.Items[0]
	publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

	return &ytChannel{
		ID:              item.ID,
		Title:           item.Snippet.Title,
		Description:     item.Snippet.Description,
		CustomURL:       item.Snippet.CustomURL,
		PublishedAt:     publishedAt,
		SubscriberCount: parseInt(item.Statistics.SubscriberCount),
		VideoCount:      parseInt(item.Statistics.VideoCount),
		ViewCount:       parseInt(item.Statistics.ViewCount),
		ThumbnailURL:    item.Snippet.Thumbnails.Default.URL,
	}, nil
}

func (d *defaultYouTubeAPI) ListVideos(ctx context.Context, channelID string, maxResults int) ([]ytVideo, error) {
	// First, search for the channel's videos
	searchParams := url.Values{
		"part":       {"snippet"},
		"channelId":  {channelID},
		"maxResults": {fmt.Sprintf("%d", maxResults)},
		"order":      {"date"},
		"type":       {"video"},
	}

	searchBody, err := d.doRequest(ctx, "search", searchParams)
	if err != nil {
		return nil, fmt.Errorf("searching videos: %w", err)
	}

	var searchResp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				Description  string `json:"description"`
				PublishedAt  string `json:"publishedAt"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   struct {
					Medium struct {
						URL string `json:"url"`
					} `json:"medium"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.Unmarshal(searchBody, &searchResp); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	if len(searchResp.Items) == 0 {
		return nil, nil
	}

	// Collect video IDs for statistics lookup
	var videoIDs string
	for i, item := range searchResp.Items {
		if i > 0 {
			videoIDs += ","
		}
		videoIDs += item.ID.VideoID
	}

	// Fetch video statistics
	statsParams := url.Values{
		"part": {"statistics"},
		"id":   {videoIDs},
	}

	statsBody, err := d.doRequest(ctx, "videos", statsParams)
	if err != nil {
		return nil, fmt.Errorf("fetching video statistics: %w", err)
	}

	var statsResp struct {
		Items []struct {
			ID         string `json:"id"`
			Statistics struct {
				ViewCount    string `json:"viewCount"`
				LikeCount    string `json:"likeCount"`
				CommentCount string `json:"commentCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.Unmarshal(statsBody, &statsResp); err != nil {
		return nil, fmt.Errorf("parsing stats response: %w", err)
	}

	// Build stats map
	statsMap := make(map[string]struct {
		Views    int
		Likes    int
		Comments int
	})
	for _, item := range statsResp.Items {
		statsMap[item.ID] = struct {
			Views    int
			Likes    int
			Comments int
		}{
			Views:    parseInt(item.Statistics.ViewCount),
			Likes:    parseInt(item.Statistics.LikeCount),
			Comments: parseInt(item.Statistics.CommentCount),
		}
	}

	// Combine search results with statistics
	var videos []ytVideo
	for _, item := range searchResp.Items {
		publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		stats := statsMap[item.ID.VideoID]

		videos = append(videos, ytVideo{
			ID:           item.ID.VideoID,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			PublishedAt:  publishedAt,
			ViewCount:    stats.Views,
			LikeCount:    stats.Likes,
			CommentCount: stats.Comments,
			ThumbnailURL: item.Snippet.Thumbnails.Medium.URL,
			ChannelTitle: item.Snippet.ChannelTitle,
		})
	}

	return videos, nil
}

// YouTubeProvider detects achievements from YouTube activity.
type YouTubeProvider struct {
	clientFactory YouTubeClientFactory
}

// NewYouTubeProvider creates a YouTube provider with the default client factory.
func NewYouTubeProvider() *YouTubeProvider {
	return &YouTubeProvider{
		clientFactory: func(ctx context.Context, accessToken string) YouTubeAPI {
			return &defaultYouTubeAPI{
				accessToken: accessToken,
				httpClient:  &http.Client{Timeout: 30 * time.Second},
			}
		},
	}
}

// NewYouTubeProviderWithFactory creates a YouTube provider with a custom client factory (for testing).
func NewYouTubeProviderWithFactory(factory YouTubeClientFactory) *YouTubeProvider {
	return &YouTubeProvider{clientFactory: factory}
}

func (p *YouTubeProvider) ID() string {
	return "YOUTUBE"
}

func (p *YouTubeProvider) Sync(ctx context.Context, accessToken string, username string) ([]DetectedAchievement, error) {
	ytAPI := p.clientFactory(ctx, accessToken)

	var achievements []DetectedAchievement

	// 1. Get channel info
	channel, err := ytAPI.GetChannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching channel: %w", err)
	}

	// 2. Channel created achievement
	achievements = append(achievements, DetectedAchievement{
		Type:        "CHANNEL_CREATED",
		Title:       fmt.Sprintf("Created YouTube channel %s", channel.Title),
		Description: truncateString(channel.Description, 200),
		Metadata: map[string]interface{}{
			"channelId":       channel.ID,
			"channelTitle":    channel.Title,
			"subscriberCount": channel.SubscriberCount,
			"videoCount":      channel.VideoCount,
			"viewCount":       channel.ViewCount,
		},
		ProofURL: fmt.Sprintf("https://youtube.com/channel/%s", channel.ID),
		ProofData: map[string]interface{}{
			"channelId": channel.ID,
		},
		SourceID:   fmt.Sprintf("youtube:channel:%s", channel.ID),
		OccurredAt: channel.PublishedAt,
	})

	// 3. Subscriber milestones
	subAchievements := p.detectSubscriberMilestones(channel)
	achievements = append(achievements, subAchievements...)

	// 4. Fetch videos and detect video achievements
	videos, err := ytAPI.ListVideos(ctx, channel.ID, 50)
	if err != nil {
		// Don't fail entirely if video fetch fails
		return achievements, nil
	}

	videoAchievements := p.detectVideoAchievements(videos, channel)
	achievements = append(achievements, videoAchievements...)

	// 5. View milestones per video
	viewAchievements := p.detectViewMilestones(videos, channel)
	achievements = append(achievements, viewAchievements...)

	return achievements, nil
}

func (p *YouTubeProvider) detectSubscriberMilestones(channel *ytChannel) []DetectedAchievement {
	var achievements []DetectedAchievement

	for _, threshold := range subscriberMilestones {
		if channel.SubscriberCount >= threshold {
			achievements = append(achievements, DetectedAchievement{
				Type:        "SUBSCRIBERS_MILESTONE",
				Title:       fmt.Sprintf("%s reached %s subscribers", channel.Title, formatNumber(threshold)),
				Description: fmt.Sprintf("YouTube channel %s has reached %s subscribers", channel.Title, formatNumber(threshold)),
				Metadata: map[string]interface{}{
					"channelId":    channel.ID,
					"channelTitle": channel.Title,
					"threshold":    threshold,
					"current":      channel.SubscriberCount,
				},
				ProofURL: fmt.Sprintf("https://youtube.com/channel/%s", channel.ID),
				ProofData: map[string]interface{}{
					"channelId":       channel.ID,
					"subscriberCount": channel.SubscriberCount,
				},
				SourceID:   fmt.Sprintf("youtube:subscribers:%s:%d", channel.ID, threshold),
				OccurredAt: channel.PublishedAt, // approximate
			})
		}
	}

	return achievements
}

func (p *YouTubeProvider) detectVideoAchievements(videos []ytVideo, channel *ytChannel) []DetectedAchievement {
	var achievements []DetectedAchievement

	for _, video := range videos {
		achievements = append(achievements, DetectedAchievement{
			Type:        "VIDEO_PUBLISHED",
			Title:       fmt.Sprintf("Published \"%s\"", truncateString(video.Title, 60)),
			Description: truncateString(video.Description, 200),
			Metadata: map[string]interface{}{
				"videoId":      video.ID,
				"videoTitle":   video.Title,
				"channelTitle": channel.Title,
				"viewCount":    video.ViewCount,
				"likeCount":    video.LikeCount,
				"commentCount": video.CommentCount,
			},
			ProofURL: fmt.Sprintf("https://youtube.com/watch?v=%s", video.ID),
			ProofData: map[string]interface{}{
				"videoId": video.ID,
			},
			SourceID:   fmt.Sprintf("youtube:video:%s", video.ID),
			OccurredAt: video.PublishedAt,
		})
	}

	return achievements
}

func (p *YouTubeProvider) detectViewMilestones(videos []ytVideo, channel *ytChannel) []DetectedAchievement {
	var achievements []DetectedAchievement

	for _, video := range videos {
		for _, threshold := range viewMilestones {
			if video.ViewCount >= threshold {
				achievements = append(achievements, DetectedAchievement{
					Type:        "VIEWS_MILESTONE",
					Title:       fmt.Sprintf("\"%s\" reached %s views", truncateString(video.Title, 40), formatNumber(threshold)),
					Description: fmt.Sprintf("Video \"%s\" on YouTube has reached %s views", video.Title, formatNumber(threshold)),
					Metadata: map[string]interface{}{
						"videoId":    video.ID,
						"videoTitle": video.Title,
						"threshold":  threshold,
						"current":    video.ViewCount,
					},
					ProofURL: fmt.Sprintf("https://youtube.com/watch?v=%s", video.ID),
					ProofData: map[string]interface{}{
						"videoId":   video.ID,
						"viewCount": video.ViewCount,
					},
					SourceID:   fmt.Sprintf("youtube:views:%s:%d", video.ID, threshold),
					OccurredAt: video.PublishedAt,
				})
			}
		}
	}

	return achievements
}

// parseInt safely parses a string to int, returning 0 on failure.
func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// formatNumber returns a human-readable number string (e.g., "10K", "1M").
func formatNumber(n int) string {
	switch {
	case n >= 1000000:
		if n%1000000 == 0 {
			return fmt.Sprintf("%dM", n/1000000)
		}
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	case n >= 1000:
		if n%1000 == 0 {
			return fmt.Sprintf("%dK", n/1000)
		}
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
