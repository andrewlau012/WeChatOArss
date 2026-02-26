package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/spf13/viper"

	"wechatoarss/internal/model"
	"wechatoarss/internal/store"
)

type WechatService struct {
	request *gorequest.SuperAgent
}

func NewWechatService() *WechatService {
	return &WechatService{
		request: gorequest.New(),
	}
}

// LoginQRCode represents login QR code response
type LoginQRCode struct {
	ErrMsg     string `json:"errMsg"`
	UUID       string `json:"uuid"`
	Tips       string `json:"tips"`
	IsLogin    bool   `json:"isLogin"`
	QRCode     string `json:"qrcode"` // Base64 image
	RedirectURL string `json:"redirectUrl"`
}

// LoginStatus represents login status
type LoginStatus struct {
	ErrMsg     string `json:"errMsg"`
	Code       int    `json:"code"`
	RedirectURL string `json:"redirect_url"`
	Cookie     string `json:"cookie"`
}

// GetLoginQRCode returns login QR code for scanning
func (s *WechatService) GetLoginQRCode() (*LoginQRCode, error) {
	// Use wechat2rss's API endpoint
	resp, body, errs := s.request.Get("https://wechat2rss.xlab.app/api/login/new").
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		End()
	
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result LoginQRCode
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CheckLoginStatus checks if QR code is scanned
func (s *WechatService) CheckLoginStatus(uuid string) (*LoginStatus, error) {
	resp, body, errs := s.request.Get(fmt.Sprintf("https://wechat2rss.xlab.app/api/login/status?uuid=%s", uuid)).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		End()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result LoginStatus
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBizIDByURL gets biz_id from article URL
func (s *WechatService) GetBizIDByURL(articleURL string) (string, error) {
	// Parse URL to get biz and mid parameters
	u, err := url.Parse(articleURL)
	if err != nil {
		return "", err
	}

	queryParams := u.Query()
	biz := queryParams.Get("biz")
	mid := queryParams.Get("mid")
	sn := queryParams.Get("sn")

	if biz != "" {
		return biz, nil
	}

	// Need to fetch page to get biz_id
	resp, body, errs := s.request.Get(articleURL).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		Set("Cookie", viper.GetString("wechat.cookie")).
		End()

	if len(errs) > 0 {
		return "", errs[0]
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Extract biz_id from page
	bizRegex := regexp.MustCompile(`biz\s*=\s*["']([^"']+)["']`)
	matches := bizRegex.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("biz_id not found")
}

// GetChannelInfo gets channel info by biz_id
func (s *WechatService) GetChannelInfo(bizID string) (*model.Channel, error) {
	// Use wechat2rss API
	resp, body, errs := s.request.Get(fmt.Sprintf("https://wechat2rss.xlab.app/api/channel/%s", bizID)).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		End()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Parse response to get channel info
	// For now, create a basic channel
	channel := &model.Channel{
		BizID:       bizID,
		Name:        bizID,
		Description: "公众号",
		Status:      "active",
	}

	return channel, nil
}

// GetArticles gets articles from a channel
func (s *WechatService) GetArticles(bizID string, offset, count int) ([]model.Article, error) {
	// Use wechat2rss API
	apiURL := fmt.Sprintf("https://wechat2rss.xlab.app/api/articles?biz_id=%s&offset=%d&count=%d", bizID, offset, count)
	
	resp, body, errs := s.request.Get(apiURL).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		End()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result struct {
		Err   string        `json:"err"`
		Data  []model.Article `json:"data"`
	}

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetArticleContent gets full article content
func (s *WechatService) GetArticleContent(articleURL string) (string, error) {
	resp, body, errs := s.request.Get(articleURL).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		Set("Cookie", viper.GetString("wechat.cookie")).
		End()

	if len(errs) > 0 {
		return "", errs[0]
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Extract content from HTML
	// Look for id="js_content"
	contentRegex := regexp.MustCompile(`id=["']js_content["'][^>]*>([\s\S]*?)</div>`)
	matches := contentRegex.FindStringSubmatch(body)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	return "", nil
}

// SearchChannels searches for public accounts
func (s *WechatService) SearchChannels(keyword string) ([]model.Channel, error) {
	// Use wechat2rss API if available
	apiURL := fmt.Sprintf("https://wechat2rss.xlab.app/api/search?q=%s", url.QueryEscape(keyword))
	
	resp, body, errs := s.request.Get(apiURL).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36").
		End()

	if len(errs) > 0 {
		// Return mock data if API unavailable
		return []model.Channel{
			{
				BizID:       generateBizID(),
				Name:        keyword + "公众号",
				Description: "搜索结果 - " + keyword,
				Status:      "active",
			},
		}, nil
	}

	if resp.StatusCode != 200 {
		return []model.Channel{
			{
				BizID:       generateBizID(),
				Name:        keyword + "公众号",
				Description: "搜索结果 - " + keyword,
				Status:      "active",
			},
		}, nil
	}

	var result struct {
		Err  string          `json:"err"`
		Data []model.Channel `json:"data"`
	}

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GenerateHMAC generates HMAC for feed ID encryption
func (s *WechatService) GenerateHMAC(bizID string) string {
	secret := viper.GetString("rss.secret")
	if secret == "" {
		secret = "default_secret"
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(bizID))
	return hex.EncodeToString(h.Sum(nil))
}

// EncryptFeedID encrypts feed ID
func (s *WechatService) EncryptFeedID(bizID string) string {
	hmacStr := s.GenerateHMAC(bizID)
	// Return first 16 characters
	if len(hmacStr) > 16 {
		return hmacStr[:16]
	}
	return hmacStr
}

// DecryptFeedID decrypts feed ID
func (s *WechatService) DecryptFeedID(encryptedID string) string {
	// Since HMAC is one-way, we need to check all channels
	channels, _, err := store.GetChannels(1, 10000, "")
	if err != nil {
		return ""
	}

	for _, ch := range channels {
		if s.EncryptFeedID(ch.BizID) == encryptedID {
			return ch.BizID
		}
	}

	return ""
}

func generateBizID() string {
	rand.Seed(time.Now().UnixNano()
	// Generate random biz_id (mkt=xxx format)
	letters := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "Mzkz" + string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
