package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

	// Get executable directory for static files
	execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	webDist := filepath.Join(execDir, "web", "dist")

	// Fallback to current directory if not found
	if _, err := os.Stat(webDist); err != nil {
		webDist = "./web/dist"
	}

	log.Printf("Serving static files from: %s", webDist)

	// CORS configuration
	router.Use(corsMiddleware())

	// Serve static files
	router.Static("/assets", filepath.Join(webDist, "assets"))
	router.StaticFile("/favicon.ico", filepath.Join(webDist, "favicon.ico"))
	router.StaticFile("/favicon.svg", filepath.Join(webDist, "favicon.svg"))

	// SPA fallback - serve index.html for all other routes
	router.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(webDist, "index.html"))
	})

	// Initialize handlers
	h := handler.NewHandler(wechatSvc, fetcherSvc)

	// Login routes (no auth required) - must be before /api group
	router.GET("/login/new", h.GetLoginQRCode)
	router.POST("/login/code", h.SubmitLoginCode)
	router.GET("/login/status", h.GetLoginStatus)

	// API routes (require auth)
	api := router.Group("/api")
	api.Use(authMiddleware())
	{
		// Account management
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

	// RSS routes (public) - using query param for format
	rss := router.Group("")
	{
		rss.GET("/feed/:id", h.GetRSSFeed)
		rss.GET("/feed/all", h.GetRSSAll)
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
