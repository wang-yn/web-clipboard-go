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
	"web-clipboard-go/internal/handlers"
	"web-clipboard-go/internal/middleware"
	"web-clipboard-go/internal/models"
	"web-clipboard-go/internal/services"
)

func main() {
	// Initialize user manager
	userManager, err := services.NewUserManager("./data")
	if err != nil {
		log.Fatal("Failed to initialize user manager:", err)
	}

	app := &models.App{
		ClipboardData: make(map[string]*models.ClipboardItem),
		DataMutex:     &sync.RWMutex{},
		TempDir:       getTempDir(),
		Security:      services.NewSecurityService(),
		RateLimiter:   services.NewRateLimitService(),
		UserManager:   userManager,
		AuthService:   services.NewAuthService(userManager),
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
	}

	// Protected API endpoints (require authentication)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(app))
	{
		api.POST("/text", handler.SaveText)
		api.GET("/text/:id", handler.GetText)
		api.POST("/file", handler.SaveFile)
		api.GET("/file/:id", handler.GetFile)
		api.DELETE("/:id", handler.DeleteItem)
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
		users.PUT("/:id/password", handler.ChangeUserPassword)
	}

	// Admin-only cleanup endpoint
	api.GET("/cleanup", middleware.AdminMiddleware(app), handler.Cleanup)

	// Serve specific static files (public)
	router.GET("/app.js", func(c *gin.Context) {
		c.File("./web/static/js/app.js")
	})
	router.GET("/auth.js", func(c *gin.Context) {
		c.File("./web/static/js/auth.js")
	})
	router.GET("/i18n.js", func(c *gin.Context) {
		c.File("./web/static/js/i18n.js")
	})
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./web/static/favicon.ico")
	})

	// Public routes for login and main pages
	router.GET("/login.html", func(c *gin.Context) {
		c.File("./web/templates/login.html")
	})
	router.GET("/", func(c *gin.Context) {
		c.File("./web/templates/index.html")
	})

	return router
}

func getTempDir() string {
	return filepath.Join(os.TempDir(), "web-clipboard-go")
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
