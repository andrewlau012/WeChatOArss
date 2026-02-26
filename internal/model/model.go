package model

import "time"

// Account represents a WeChat account
type Account struct {
	ID         int64     `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Cookie     string    `json:"-" db:"cookie"`
	Token      string    `json:"-" db:"token"`
	Available  bool      `json:"available" db:"available"`
	NeedCheck  bool      `json:"needCheck" db:"need_check"`
	WaitTime   time.Time `json:"waitTime" db:"wait_time"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// Channel represents a subscribed WeChat public account
type Channel struct {
	ID           int64     `json:"id" db:"id"`
	BizID        string    `json:"biz_id" db:"biz_id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Avatar       string    `json:"avatar" db:"avatar"`
	Link         string    `json:"link" db:"link"`
	AccountID    int64     `json:"accountId" db:"account_id"`
	LastUpdate   time.Time `json:"lastUpdate" db:"last_update"`
	ArticleCount int       `json:"articleCount" db:"article_count"`
	Status       string    `json:"status" db:"status"` // active, paused
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// Article represents an article from a channel
type Article struct {
	ID          int64     `json:"id" db:"id"`
	BizID       string    `json:"biz_id" db:"biz_id"`
	ChannelName string    `json:"channelName" db:"-"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"desc" db:"description"`
	Content     string    `json:"content" db:"content"`
	Link        string    `json:"link" db:"link"`
	Cover       string    `json:"cover" db:"cover"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	PublishedAt time.Time `json:"publishedAt" db:"published_at"`
}

// RSSItem represents an item in RSS feed
type RSSItem struct {
	Title       string
	Link        string
	Description string
	Content     string
	PubDate     string
	Cover       string
}

// Config represents system configuration
type Config struct {
	Host                 string   `json:"host" yaml:"host"`
	Token               string   `json:"token" yaml:"token"`
	RSSToken            string   `json:"rssToken" yaml:"rss_token"`
	MaxItemCount        int      `json:"maxItemCount" yaml:"max_item_count"`
	KeepOldCount        int      `json:"keepOldCount" yaml:"keep_old_count"`
	EncFeedID           bool     `json:"encFeedId" yaml:"enc_feed_id"`
	Static              bool     `json:"static" yaml:"static"`
	SchedulerTimes      []string `json:"schedulerTimes" yaml:"scheduler_times"`
	NotifyEnabled       bool     `json:"notifyEnabled" yaml:"notify_enabled"`
	NotifyType          string   `json:"notifyType" yaml:"notify_type"`
	TelegramToken       string   `json:"telegramToken" yaml:"telegram_token"`
	TelegramAdminUID    string   `json:"telegramAdminUid" yaml:"telegram_admin_uid"`
	ServerChanKey       string   `json:"serverChanKey" yaml:"server_chan_key"`
	WebhookURL          string   `json:"webhookUrl" yaml:"webhook_url"`
	BarkURL             string   `json:"barkUrl" yaml:"bark_url"`
}

// LoginResponse represents login QR code response
type LoginResponse struct {
	IsLogin bool   `json:"isLogin"`
	QRCode  string `json:"qrcode"`
}

// LoginResult represents login result
type LoginResult struct {
	Err   string `json:"err"`
	Tips  string `json:"tips"`
	RedirURL string `json:"redir_url"`
}

// APIResponse represents standard API response
type APIResponse struct {
	Err   string      `json:"err"`
	Data  interface{} `json:"data,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
}
