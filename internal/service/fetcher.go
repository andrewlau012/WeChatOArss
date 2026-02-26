package service

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"

	"wechatoarss/internal/model"
	"wechatoarss/internal/store"
)

type FetcherService struct {
	wechatSvc *WechatService
}

func NewFetcherService(wechatSvc *WechatService) *FetcherService {
	return &FetcherService{
		wechatSvc: wechatSvc,
	}
}

// FetchChannel fetches articles for a channel
func (s *FetcherService) FetchChannel(bizID string) error {
	log.Printf("Fetching channel: %s", bizID)

	// Get channel info
	ch, err := store.GetChannelByBizID(bizID)
	if err != nil {
		return err
	}
	_ = ch // use the channel

	// Get articles
	articles, err := s.wechatSvc.GetArticles(bizID, 0, 20)
	if err != nil {
		log.Printf("Failed to get articles for %s: %v", bizID, err)
		return err
	}

	// Save articles
	count := 0
	for _, article := range articles {
		// Parse published time
		publishedAt := article.PublishedAt
		if publishedAt.IsZero() {
			publishedAt = time.Now()
		}

		// Get full content if needed
		content := article.Content
		if content == "" {
			content, _ = s.wechatSvc.GetArticleContent(article.Link)
		}

		_, err := store.CreateArticle(
			bizID,
			article.Title,
			article.Description,
			content,
			article.Link,
			article.Cover,
			publishedAt,
		)
		if err != nil {
			continue
		}
		count++
	}

	// Update channel article count
	store.UpdateChannelArticleCount(bizID, count)
	log.Printf("Fetched %d articles for channel %s", count, bizID)

	return nil
}

// FetchAll fetches all active channels
func (s *FetcherService) FetchAll() error {
	channels, err := store.GetActiveChannels()
	if err != nil {
		return err
	}

	log.Printf("Fetching %d channels", len(channels))

	for _, channel := range channels {
		if err := s.FetchChannel(channel.BizID); err != nil {
			log.Printf("Error fetching channel %s: %v", channel.BizID, err)
			continue
		}

		// Small delay between channels
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// AddChannel adds a new channel
func (s *FetcherService) AddChannel(bizID string) (string, error) {
	// Check if already exists
	existing, err := store.GetChannelByBizID(bizID)
	if err == nil && existing != nil {
		// Trigger update
		go s.FetchChannel(bizID)
		return existing.Link, nil
	}

	// Get channel info
	channel, err := s.wechatSvc.GetChannelInfo(bizID)
	if err != nil {
		return "", err
	}

	// Create channel
	newChannel, err := store.CreateChannel(
		channel.BizID,
		channel.Name,
		channel.Description,
		channel.Avatar,
		channel.Link,
		1, // Default account
	)
	if err != nil {
		return "", err
	}

	// Trigger first fetch
	go s.FetchChannel(bizID)

	return newChannel.Link, nil
}

// AddChannelByURL adds channel by article URL
func (s *FetcherService) AddChannelByURL(articleURL string) (string, error) {
	// Extract biz_id from URL
	bizID, err := s.wechatSvc.GetBizIDByURL(articleURL)
	if err != nil {
		return "", err
	}

	return s.AddChannel(bizID)
}

// ParseArticleContent parses article HTML content
func (s *FetcherService) ParseArticleContent(html string) string {
	// Remove script tags
	re := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	html = re.ReplaceAllString(html, "")

	// Remove style tags
	re = regexp.MustCompile(`<style[^>]*>[\s\S]*?</style>`)
	html = re.ReplaceAllString(html, "")

	// Clean up
	html = strings.TrimSpace(html)

	return html
}

// FormatArticleDate formats date for RSS
func (s *FetcherService) FormatArticleDate(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	return t.Format("Mon, 02 Jan 2006 15:04:05 -0700")
}

// ExtractImages extracts image URLs from content
func (s *FetcherService) ExtractImages(content string) []string {
	re := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(content, -1)

	var images []string
	for _, match := range matches {
		if len(match) > 1 {
			images = append(images, match[1])
		}
	}

	return images
}

// CleanDescription cleans article description
func (s *FetcherService) CleanDescription(desc string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]+>`)
	desc = re.ReplaceAllString(desc, "")

	// Decode HTML entities
	desc = strings.ReplaceAll(desc, "&nbsp;", " ")
	desc = strings.ReplaceAll(desc, "&amp;", "&")
	desc = strings.ReplaceAll(desc, "&lt;", "<")
	desc = strings.ReplaceAll(desc, "&gt;", ">")
	desc = strings.ReplaceAll(desc, "&quot;", `"`)
	desc = strings.ReplaceAll(desc, "&#39;", "'")

	// Trim whitespace
	desc = strings.TrimSpace(desc)

	// Limit length
	if len(desc) > 200 {
		desc = desc[:200] + "..."
	}

	return desc
}

// GetArticleCount gets total article count for a channel
func (s *FetcherService) GetArticleCount(bizID string) (int, error) {
	_, total, err := store.GetArticles(bizID, "", "", 1, 1, false)
	return total, err
}

// SearchChannels searches channels
func (s *FetcherService) SearchChannels(keyword string) ([]model.Channel, error) {
	channels, total, err := store.GetChannels(1, 20, keyword)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		// Try search via wechat API
		return s.wechatSvc.SearchChannels(keyword)
	}

	return channels, nil
}

// DeleteChannel deletes a channel
func (s *FetcherService) DeleteChannel(bizID string) error {
	return store.DeleteChannel(bizID)
}

// PauseChannel pauses a channel
func (s *FetcherService) PauseChannel(bizID string, pause bool) error {
	status := "active"
	if pause {
		status = "paused"
	}
	return store.UpdateChannelStatus(bizID, status)
}

// ParseBizID parses biz_id from string (handles both plain and encrypted)
func (s *FetcherService) ParseBizID(id string) string {
	// If enc_feed_id is enabled, id might be encrypted
	if viper.GetBool("rss.enc_feed_id") {
		return s.wechatSvc.DecryptFeedID(id)
	}
	return id
}

// IsValidBizID checks if biz_id is valid
func (s *FetcherService) IsValidBizID(bizID string) bool {
	_, err := store.GetChannelByBizID(bizID)
	return err == nil
}
