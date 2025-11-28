package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize user manager
	userManager, err := NewUserManager("./data")
	if err != nil {
		log.Fatal("Failed to initialize user manager:", err)
	}

	app := &App{
		clipboardData: make(map[string]*ClipboardItem),
		dataMutex:     &sync.RWMutex{},
		tempDir:       getTempDir(),
		security:      NewSecurityService(),
		rateLimiter:   NewRateLimitService(),
		userManager:   userManager,
		authService:   NewAuthService(userManager),
	}

	app.initTempDir()

	server := &http.Server{
		Addr:         ":5000",
		Handler:      app.setupRouter(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Starting server on :5000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	app.startCleanupService()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	app.stopCleanupService()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	fmt.Println("Server exiting")
}

func (app *App) setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.Use(app.corsMiddleware())
	router.Use(app.securityHeadersMiddleware())
	router.Use(app.rateLimitMiddleware())

	// Public auth endpoints
	auth := router.Group("/api/auth")
	{
		auth.POST("/login", app.login)
		auth.POST("/logout", app.authMiddleware(), app.logout)
		auth.GET("/me", app.authMiddleware(), app.getCurrentUser)
	}

	// Protected API endpoints (require authentication)
	api := router.Group("/api")
	api.Use(app.authMiddleware())
	{
		api.POST("/text", app.saveText)
		api.GET("/text/:id", app.getText)
		api.POST("/file", app.saveFile)
		api.GET("/file/:id", app.getFile)
		api.DELETE("/:id", app.deleteItem)
	}

	// Admin-only endpoints
	users := api.Group("/users")
	users.Use(app.adminMiddleware())
	{
		users.POST("", app.createUser)
		users.GET("", app.listUsers)
		users.GET("/:id", app.getUser)
		users.PUT("/:id", app.updateUser)
		users.DELETE("/:id", app.deleteUser)
		users.PUT("/:id/password", app.changeUserPassword)
	}

	// Admin-only cleanup endpoint
	api.GET("/cleanup", app.adminMiddleware(), app.cleanup)

	// Serve specific static files (public)
	router.GET("/app.js", func(c *gin.Context) {
		c.File("./wwwroot/app.js")
	})
	router.GET("/auth.js", func(c *gin.Context) {
		c.File("./wwwroot/auth.js")
	})
	router.GET("/i18n.js", func(c *gin.Context) {
		c.File("./wwwroot/i18n.js")
	})
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./wwwroot/favicon.ico")
	})

	// Public routes for login and main pages
	router.GET("/login.html", func(c *gin.Context) {
		c.File("./wwwroot/login.html")
	})
	router.GET("/", func(c *gin.Context) {
		c.File("./wwwroot/index.html")
	})

	return router
}

func (app *App) startCleanupService() {
	app.cleanupTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for range app.cleanupTicker.C {
			app.performCleanup()
		}
	}()
}

func (app *App) stopCleanupService() {
	if app.cleanupTicker != nil {
		app.cleanupTicker.Stop()
	}
}

func (app *App) performCleanup() {
	removedCount := 0
	now := time.Now().UTC()

	app.dataMutex.Lock()
	for id, item := range app.clipboardData {
		if item.ExpiresAt.Before(now) {
			if item.Type == "file" && item.FilePath != "" {
				os.Remove(item.FilePath)
			}
			delete(app.clipboardData, id)
			removedCount++
		}
	}
	app.dataMutex.Unlock()

	if removedCount > 0 {
		fmt.Printf("Cleaned up %d expired items\n", removedCount)
	}

	app.security.CleanupExpired()
	app.rateLimiter.CleanupExpired()
	app.authService.CleanupExpiredSessions()
}
