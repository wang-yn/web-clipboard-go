package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *App) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigins := []string{
			"http://localhost:5000",
			"https://localhost:5001",
			"http://localhost:8080",
			"https://localhost:8443",
			"http://127.0.0.1:5000",
			"http://127.0.0.1:8080",
		}

		origin := c.GetHeader("Origin")
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Disposition")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (app *App) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self' 'unsafe-inline' cdn.tailwindcss.com")
		c.Next()
	}
}

func (app *App) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		ip := app.security.getClientIP(c)
		endpoint := c.Request.Method

		if !app.rateLimiter.IsAllowed(ip, endpoint) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded. Please slow down."})
			c.Abort()
			return
		}

		c.Next()
	}
}