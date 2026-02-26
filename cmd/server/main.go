package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"wechatoarss/internal/config"
	"wechatoarss/internal/handler"
	"wechatoarss/internal/service"
	"wechatoarss/internal/store"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := store.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	wechatSvc := service.NewWechatService()
	fetcherSvc := service.NewFetcherService(wechatSvc)
	schedulerSvc := service.NewSchedulerService(fetcherSvc)

	// Start scheduler
	if err := schedulerSvc.Start(); err != nil {
		log.Printf("Warning: Failed to start scheduler: %v", err)
	}

	// Setup router
	router := setupRouter(wechatSvc, fetcherSvc)

	// Start server
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("WeChatOArss server started on port %s", port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop scheduler
	schedulerSvc.Stop()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(wechatSvc *service.WechatService, fetcherSvc *service.FetcherService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	router.Use(corsMiddleware())

	// Serve static files
	router.Static("/static", "./web/dist/static")
	router.StaticFile("/", "./web/dist/index.html")
	router.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// Initialize handlers
	h := handler.NewHandler(wechatSvc, fetcherSvc)

	// API routes (require auth)
	api := router.Group("/api")
	api.Use(authMiddleware())
	{
		// Account management
		api.GET("/login/new", h.GetLoginQRCode)
		api.POST("/login/code", h.SubmitLoginCode)
		api.GET("/login/list", h.ListAccounts)
		api.POST("/login/refresh/:id", h.RefreshAccountStatus)
		api.DELETE("/login/del/:id", h.DeleteAccount)

		// Channel management
		api.GET("/add/:id", h.AddChannel)
		api.GET("/addurl", h.AddChannelByURL)
		api.DELETE("/del/:id", h.DeleteChannel)
		api.GET("/pause/:id", h.PauseChannel)
		api.GET("/list", h.ListChannels)

		// Articles
		api.GET("/query", h.QueryArticles)
		api.GET("/article/:id", h.GetArticle)

		// Config
		api.GET("/config", h.GetConfig)
		api.POST("/config", h.UpdateConfig)

		// Export
		api.GET("/opml", h.ExportOPML)
	}

	// RSS routes (public)
	rss := router.Group("")
	{
		rss.GET("/feed/:id.xml", h.GetRSSFeed)
		rss.GET("/feed/:id.json", h.GetRSSFeedJSON)
		rss.GET("/feed/all.xml", h.GetRSSAll)
		rss.GET("/feed/all.json", h.GetRSSAllJSON)
	}

	// Protected RSS routes
	protectedRss := router.Group("/feed")
	protectedRss.Use(authMiddleware())
	{
		protectedRss.GET("/all.xml", h.GetRSSAll)
		protectedRss.GET("/all.json", h.GetRSSAllJSON)
	}

	// Proxy routes
	proxy := router.Group("")
	proxy.Use(authMiddleware())
	{
		proxy.GET("/img-proxy", h.ImageProxy)
		proxy.GET("/video-proxy", h.VideoProxy)
		proxy.GET("/link-proxy", h.LinkProxy)
	}

	// System routes
	router.GET("/version", h.GetVersion)

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("k")
		expectedToken := viper.GetString("server.token")

		if expectedToken != "" && token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"err": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
