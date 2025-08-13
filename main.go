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
)

func main() {
	app := &App{
		clipboardData: make(map[string]*ClipboardItem),
		tempDir:       getTempDir(),
		security:      NewSecurityService(),
		rateLimiter:   NewRateLimitService(),
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

	api := router.Group("/api")
	{
		api.POST("/text", app.saveText)
		api.GET("/text/:id", app.getText)
		api.POST("/file", app.saveFile)
		api.GET("/file/:id", app.getFile)
		api.DELETE("/:id", app.deleteItem)
		api.GET("/cleanup", app.cleanup)
	}

	// Serve specific static files
	router.GET("/app.js", func(c *gin.Context) {
		c.File("./wwwroot/app.js")
	})
	router.GET("/i18n.js", func(c *gin.Context) {
		c.File("./wwwroot/i18n.js")
	})
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./wwwroot/favicon.ico")
	})

	// Default route serves index.html
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
}
