package handler

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"wechatoarss/internal/model"
	"wechatoarss/internal/service"
	"wechatoarss/internal/store"
)

type Handler struct {
	wechatSvc  *service.WechatService
	fetcherSvc *service.FetcherService
}

func NewHandler(wechatSvc *service.WechatService, fetcherSvc *service.FetcherService) *Handler {
	return &Handler{
		wechatSvc:  wechatSvc,
		fetcherSvc: fetcherSvc,
	}
}

// Wechat login handlers
func (h *Handler) GetLoginQRCode(c *gin.Context) {
	qrCode, err := h.wechatSvc.GetLoginQRCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"err": "",
		"data": gin.H{
			"isLogin": qrCode.IsLogin,
			"qrcode":  qrCode.QRCode,
			"uuid":    qrCode.UUID,
		},
	})
}

func (h *Handler) GetLoginStatus(c *gin.Context) {
	uuid := c.Query("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "uuid is required"})
		return
	}

	status, err := h.wechatSvc.CheckLoginStatus(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"err":    "",
		"code":   status.Code,
		"redir_url": status.RedirectURL,
	})
}

func (h *Handler) SubmitLoginCode(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: err.Error()})
		return
	}

	// In a real implementation, this would verify the code
	c.JSON(http.StatusOK, model.APIResponse{Err: ""})
}

func (h *Handler) ListAccounts(c *gin.Context) {
	accounts, err := store.GetAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"err":  "",
		"data": accounts,
	})
}

func (h *Handler) RefreshAccountStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid ID"})
		return
	}

	// Refresh account status
	account, err := store.GetAccountByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{Err: "Account not found"})
		return
	}

	account.Available = true
	account.NeedCheck = false

	c.JSON(http.StatusOK, model.APIResponse{Err: ""})
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid ID"})
		return
	}

	if err := store.DeleteAccount(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Err: ""})
}

// Channel handlers
func (h *Handler) AddChannel(c *gin.Context) {
	bizID := c.Param("id")
	if bizID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid biz_id"})
		return
	}

	_, err := h.fetcherSvc.AddChannel(bizID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	c.JSON(http.StatusOK, gin.H{
		"err":  "",
		"data": fmt.Sprintf("%s/feed/%s.xml", host, bizID),
	})
}

func (h *Handler) AddChannelByURL(c *gin.Context) {
	articleURL := c.Query("url")
	if articleURL == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "url is required"})
		return
	}

	_, err := h.fetcherSvc.AddChannelByURL(articleURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	// Extract biz_id from URL
	bizID, _ := h.wechatSvc.GetBizIDByURL(articleURL)

	c.JSON(http.StatusOK, gin.H{
		"err":  "",
		"data": fmt.Sprintf("%s/feed/%s.xml", host, bizID),
	})
}

func (h *Handler) DeleteChannel(c *gin.Context) {
	bizID := c.Param("id")
	if bizID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid biz_id"})
		return
	}

	// Parse biz_id (handle encrypted)
	bizID = h.fetcherSvc.ParseBizID(bizID)

	if err := h.fetcherSvc.DeleteChannel(bizID); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Err: ""})
}

func (h *Handler) PauseChannel(c *gin.Context) {
	bizID := c.Param("id")
	if bizID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid biz_id"})
		return
	}

	// Parse biz_id (handle encrypted)
	bizID = h.fetcherSvc.ParseBizID(bizID)

	status := c.DefaultQuery("status", "false")
	pause := status == "true"

	if err := h.fetcherSvc.PauseChannel(bizID, pause); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{Err: ""})
}

func (h *Handler) ListChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	name := c.Query("name")

	channels, total, err := store.GetChannels(page, size, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	// Build response
	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	var data []gin.H
	for _, ch := range channels {
		feedID := ch.BizID
		if viper.GetBool("rss.enc_feed_id") {
			feedID = h.wechatSvc.EncryptFeedID(ch.BizID)
		}

		data = append(data, gin.H{
			"id":           ch.ID,
			"biz_id":       ch.BizID,
			"name":         ch.Name,
			"description":  ch.Description,
			"avatar":       ch.Avatar,
			"link":         fmt.Sprintf("%s/feed/%s.xml", host, feedID),
			"lastUpdate":   ch.LastUpdate.Format("2006-01-02 15:04:05"),
			"articleCount": ch.ArticleCount,
			"status":       ch.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"err": "",
		"data": data,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// Article handlers
func (h *Handler) QueryArticles(c *gin.Context) {
	bizID := c.Query("bid")
	before := c.Query("before")
	after := c.Query("after")
	content := c.DefaultQuery("content", "1")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	// Handle encrypted biz_id
	if bizID != "" && viper.GetBool("rss.enc_feed_id") {
		bizID = h.fetcherSvc.ParseBizID(bizID)
	}

	includeContent := content == "1"

	articles, total, err := store.GetArticles(bizID, before, after, page, size, includeContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	// Add channel names
	for i := range articles {
		ch, err := store.GetChannelByBizID(articles[i].BizID)
		if err == nil && ch != nil {
			articles[i].ChannelName = ch.Name
		}
	}

	var data []gin.H
	for _, a := range articles {
		item := gin.H{
			"biz_id":     a.BizID,
			"biz_name":   a.ChannelName,
			"title":      a.Title,
			"desc":       a.Description,
			"created":    a.PublishedAt.Format(time.RFC3339),
			"link":       a.Link,
			"cover":      a.Cover,
		}
		if includeContent {
			item["content"] = a.Content
		}
		data = append(data, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"err": "",
		"data": data,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

func (h *Handler) GetArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid article ID"})
		return
	}

	article, err := store.GetArticleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{Err: "Article not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"err": "",
		"data": gin.H{
			"id":           article.ID,
			"biz_id":       article.BizID,
			"biz_name":     article.ChannelName,
			"title":        article.Title,
			"desc":         article.Description,
			"content":      article.Content,
			"created":      article.PublishedAt.Format(time.RFC3339),
			"link":         article.Link,
			"cover":        article.Cover,
		},
	})
}

// Config handlers
func (h *Handler) GetConfig(c *gin.Context) {
	config := model.Config{
		Host:            viper.GetString("rss.host"),
		Token:           viper.GetString("server.token"),
		MaxItemCount:    viper.GetInt("rss.max_item_count"),
		KeepOldCount:    viper.GetInt("rss.keep_old_count"),
		EncFeedID:       viper.GetBool("rss.enc_feed_id"),
		Static:          viper.GetBool("rss.static"),
		SchedulerTimes:  viper.GetStringSlice("scheduler.times"),
		NotifyEnabled:   viper.GetBool("notify.enabled"),
		NotifyType:      viper.GetString("notify.type"),
	}

	c.JSON(http.StatusOK, gin.H{
		"err":  "",
		"data": config,
	})
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	var config model.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: err.Error()})
		return
	}

	if config.Host != "" {
		viper.Set("rss.host", config.Host)
	}
	if config.Token != "" {
		viper.Set("server.token", config.Token)
	}
	if config.MaxItemCount > 0 {
		viper.Set("rss.max_item_count", config.MaxItemCount)
	}
	if config.KeepOldCount != 0 {
		viper.Set("rss.keep_old_count", config.KeepOldCount)
	}
	if config.SchedulerTimes != nil {
		viper.Set("scheduler.times", config.SchedulerTimes)
	}

	c.JSON(http.StatusOK, model.APIResponse{Err: ""})
}

func (h *Handler) ExportOPML(c *gin.Context) {
	channels, _, err := store.GetChannels(1, 1000, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	var items []string
	for _, ch := range channels {
		feedID := ch.BizID
		if viper.GetBool("rss.enc_feed_id") {
			feedID = h.wechatSvc.EncryptFeedID(ch.BizID)
		}
		items = append(items, fmt.Sprintf(`<outline text="%s" title="%s" type="rss" xmlUrl="%s/feed/%s.xml"/>`, 
			html.EscapeString(ch.Name), 
			html.EscapeString(ch.Name),
			host,
			feedID))
	}

	opml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
<head>
<title>WeChatOArss Subscriptions</title>
</head>
<body>
%s
</body>
</opml>`, strings.Join(items, "\n"))

	c.Header("Content-Type", "application/xml")
	c.Header("Content-Disposition", "attachment; filename=wechatoarss.opml")
	c.String(http.StatusOK, opml)
}

func (h *Handler) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":     "v1.0.0",
		"build_time":  time.Now().Format("2006-01-02"),
		"wechatoarss": "WeChatOArss",
	})
}

// RSS handlers
func (h *Handler) GetRSSFeed(c *gin.Context) {
	bizID := c.Param("id")
	if bizID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid ID"})
		return
	}

	// Check format from path extension
	format := "xml"
	if strings.HasSuffix(bizID, ".json") {
		format = "json"
		bizID = strings.TrimSuffix(bizID, ".json")
	} else if strings.HasSuffix(bizID, ".xml") {
		bizID = strings.TrimSuffix(bizID, ".xml")
	}

	// Parse biz_id (handle encrypted)
	bizID = h.fetcherSvc.ParseBizID(bizID)

	channel, err := store.GetChannelByBizID(bizID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{Err: "Channel not found"})
		return
	}

	maxItems := viper.GetInt("rss.max_item_count")
	if maxItems == 0 {
		maxItems = 20
	}

	articles, _, err := store.GetArticles(bizID, "", "", 1, maxItems, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	// Return JSON if requested
	if format == "json" {
		jsonFeed := h.buildJSONFeed(channel.Name, channel.Description, channel.Link, host+"/feed/"+bizID+".json", articles)
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusOK, jsonFeed)
		return
	}

	rss := h.buildRSS(channel.Name, channel.Description, channel.Link, host+"/feed/"+bizID, articles, host)

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, rss)
}

func (h *Handler) GetRSSFeedJSON(c *gin.Context) {
	bizID := c.Param("id")
	if bizID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid ID"})
		return
	}

	// Parse biz_id (handle encrypted)
	bizID = h.fetcherSvc.ParseBizID(bizID)

	channel, err := store.GetChannelByBizID(bizID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{Err: "Channel not found"})
		return
	}

	maxItems := viper.GetInt("rss.max_item_count")
	if maxItems == 0 {
		maxItems = 20
	}

	articles, _, err := store.GetArticles(bizID, "", "", 1, maxItems, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	jsonFeed := h.buildJSONFeed(channel.Name, channel.Description, channel.Link, host+"/feed/"+bizID+".json", articles)

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, jsonFeed)
}

func (h *Handler) GetRSSAll(c *gin.Context) {
	// Check format from path
	path := c.Request.URL.Path
	format := "xml"
	if strings.HasSuffix(path, ".json") {
		format = "json"
	}

	maxItems := viper.GetInt("rss.max_item_count")
	if maxItems == 0 {
		maxItems = 50
	}

	articles, _, err := store.GetArticles("", "", "", 1, maxItems, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	// Get channel names
	for i := range articles {
		ch, err := store.GetChannelByBizID(articles[i].BizID)
		if err == nil && ch != nil {
			articles[i].ChannelName = ch.Name
		}
	}

	// Return JSON if requested
	if format == "json" {
		jsonFeed := h.buildJSONFeed("WeChatOArss All", "All subscribed channels", "", host+"/feed/all.json", articles)
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusOK, jsonFeed)
		return
	}

	rss := h.buildRSS("WeChatOArss All", "All subscribed channels", "", host+"/feed/all.xml", articles, host)

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, rss)
}

func (h *Handler) GetRSSAllJSON(c *gin.Context) {
	maxItems := viper.GetInt("rss.max_item_count")
	if maxItems == 0 {
		maxItems = 50
	}

	articles, _, err := store.GetArticles("", "", "", 1, maxItems, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Err: err.Error()})
		return
	}

	host := viper.GetString("rss.host")
	if host == "" {
		host = "http://localhost:8080"
	}

	// Get channel names
	for i := range articles {
		ch, err := store.GetChannelByBizID(articles[i].BizID)
		if err == nil && ch != nil {
			articles[i].ChannelName = ch.Name
		}
	}

	jsonFeed := h.buildJSONFeed("WeChatOArss All", "All subscribed channels", "", host+"/feed/all.json", articles)

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, jsonFeed)
}

// Proxy handlers
func (h *Handler) ImageProxy(c *gin.Context) {
	imgURL := c.Query("u")
	if imgURL == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "url is required"})
		return
	}

	// Decode URL
	decodedURL, err := url.QueryUnescape(imgURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid url"})
		return
	}

	// Fetch image
	resp, err := http.Get(decodedURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIResponse{Err: err.Error()})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Cache-Control", "public, max-age=86400")
	io.Copy(c.Writer, resp.Body)
}

func (h *Handler) VideoProxy(c *gin.Context) {
	h.ImageProxy(c) // Same logic
}

func (h *Handler) LinkProxy(c *gin.Context) {
	link := c.Query("u")
	if link == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "url is required"})
		return
	}

	decodedURL, err := url.QueryUnescape(link)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Err: "Invalid url"})
		return
	}

	c.Redirect(http.StatusFound, decodedURL)
}

// Helper functions
func (h *Handler) buildRSS(title, description, link, feedURL string, articles []model.Article, host string) string {
	var items []string
	for _, a := range articles {
		pubDate := a.PublishedAt.Format(time.RFC1123)
		if a.PublishedAt.IsZero() {
			pubDate = time.Now().Format(time.RFC1123)
		}

		desc := h.fetcherSvc.CleanDescription(a.Description)
		content := a.Content

		item := fmt.Sprintf(`<item>
<title><![CDATA[%s]]></title>
<link>%s</link>
<description><![CDATA[%s]]></description>
<pubDate>%s</pubDate>
<content:encoded><![CDATA[%s]]></content:encoded>
</item>`, a.Title, a.Link, desc, pubDate, content)

		items = append(items, item)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
<channel>
<title><![CDATA[%s]]></title>
<link>%s</link>
<description><![CDATA[%s]]></description>
<language>zh-cn</language>
<lastBuildDate>%s</lastBuildDate>
%s
</channel>
</rss>`, title, link, description, time.Now().Format(time.RFC1123), strings.Join(items, "\n"))
}

func (h *Handler) buildJSONFeed(title, description, homePage, feedURL string, articles []model.Article) gin.H {
	var items []gin.H
	for _, a := range articles {
		pubDate := a.PublishedAt.Format(time.RFC3339)
		if a.PublishedAt.IsZero() {
			pubDate = time.Now().Format(time.RFC3339)
		}

		items = append(items, gin.H{
			"id":           fmt.Sprintf("%d", a.ID),
			"url":          a.Link,
			"title":        a.Title,
			"content_html": a.Content,
			"summary":      a.Description,
			"date_published": pubDate,
		})
	}

	return gin.H{
		"version":       "https://jsonfeed.org/version/1",
		"title":         title,
		"home_page_url": homePage,
		"feed_url":      feedURL,
		"items":         items,
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// XML feed struct for RSS 2.0
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel RSSChannel
}

type RSSChannel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Language    string   `xml:"language"`
	LastBuildDate string  `xml:"lastBuildDate"`
	Items       []RSSItem
}

type RSSItem struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	Content     string   `xml:"content:encoded"`
}
