package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/backend/internal/handlers"
	"web-clipboard-go/backend/internal/middleware"
	"web-clipboard-go/backend/internal/models"
	"web-clipboard-go/backend/internal/services"
)

func main() {
	// Initialize user manager
	userManager, err := services.NewUserManager(getDataDir())
	if err != nil {
		log.Fatal("Failed to initialize user manager:", err)
	}
	authService := services.NewAuthService(userManager)

	app := &models.App{
		ClipboardData: make(map[string]*models.ClipboardItem),
		DataMutex:     &sync.RWMutex{},
		TempDir:       getTempDir(),
		Security:      services.NewSecurityService(),
		RateLimiter:   services.NewRateLimitService(),
		UserManager:   userManager,
		AuthService:   authService,
		OAuthService:  services.NewOAuthServiceFromEnv(userManager, authService),
	}

	initTempDir(app.TempDir)

	server := &http.Server{
		Addr:         ":5000",
		Handler:      setupRouter(app),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Starting server on :5000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	startCleanupService(app)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	stopCleanupService(app)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	fmt.Println("Server exiting")
}

func setupRouter(app *models.App) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.Use(middleware.CorsMiddleware(app))
	router.Use(middleware.SecurityHeadersMiddleware(app))
	router.Use(middleware.RateLimitMiddleware(app))

	// Create handler
	handler := &handlers.Handler{App: app}

	// Public auth endpoints
	auth := router.Group("/api/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/logout", middleware.AuthMiddleware(app), handler.Logout)
		auth.GET("/me", middleware.AuthMiddleware(app), handler.GetCurrentUser)
		auth.GET("/providers", handler.ListAuthProviders)
		auth.GET("/oauth/:provider/start", handler.StartOAuthLogin)
		auth.GET("/oauth/:provider/callback", handler.HandleOAuthCallback)
		auth.POST("/oauth/complete", handler.CompleteOAuthLogin)
	}

	// Protected API endpoints (require authentication)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(app))
	{
		api.POST("/text", handler.SaveText)
		api.GET("/text/:id", handler.GetText)
		api.POST("/file", handler.SaveFile)
		api.GET("/file/:id", handler.GetFile)
		api.GET("/items", handler.ListRecentItems)
		api.DELETE("/:id", handler.DeleteItem)
		api.PUT("/users/:id/password", handler.ChangeUserPassword)
	}

	// Admin-only endpoints
	users := api.Group("/users")
	users.Use(middleware.AdminMiddleware(app))
	{
		users.POST("", handler.CreateUser)
		users.GET("", handler.ListUsers)
		users.GET("/:id", handler.GetUser)
		users.PUT("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
	}

	// Admin-only cleanup endpoint
	api.GET("/cleanup", middleware.AdminMiddleware(app), handler.Cleanup)

	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// Public routes for login and main pages
	router.GET("/login.html", func(c *gin.Context) {
		c.File("./frontend/dist/login.html")
	})
	router.GET("/settings.html", func(c *gin.Context) {
		c.File("./frontend/dist/settings.html")
	})
	router.GET("/", func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	return router
}

func getTempDir() string {
	return filepath.Join(os.TempDir(), "web-clipboard-go")
}

func getDataDir() string {
	if value := os.Getenv("WEB_CLIPBOARD_DATA_DIR"); value != "" {
		return value
	}
	return "/data"
}

func initTempDir(tempDir string) {
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		panic("Failed to create temp directory: " + err.Error())
	}

	files, err := filepath.Glob(filepath.Join(tempDir, "*"))
	if err == nil {
		for _, file := range files {
			os.Remove(file)
		}
	}
}

func startCleanupService(app *models.App) {
	app.CleanupTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for range app.CleanupTicker.C {
			performCleanup(app)
		}
	}()
}

func stopCleanupService(app *models.App) {
	if app.CleanupTicker != nil {
		app.CleanupTicker.Stop()
	}
}

func performCleanup(app *models.App) {
	removedCount := 0
	now := time.Now().UTC()

	app.DataMutex.Lock()
	for id, item := range app.ClipboardData {
		if item.ExpiresAt.Before(now) {
			if item.Type == "file" && item.FilePath != "" {
				os.Remove(item.FilePath)
			}
			delete(app.ClipboardData, id)
			removedCount++
		}
	}
	app.DataMutex.Unlock()

	if removedCount > 0 {
		fmt.Printf("Cleaned up %d expired items\n", removedCount)
	}

	app.Security.CleanupExpired()
	app.RateLimiter.CleanupExpired()
	app.AuthService.CleanupExpiredSessions()
}
