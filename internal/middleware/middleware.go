package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/internal/models"
)

// CorsMiddleware handles CORS for the application
func CorsMiddleware(app *models.App) gin.HandlerFunc {
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

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(app *models.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self' 'unsafe-inline' cdn.tailwindcss.com")
		c.Next()
	}
}

// RateLimitMiddleware applies rate limiting to requests
func RateLimitMiddleware(app *models.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		endpoint := c.Request.Method

		if !app.RateLimiter.IsAllowed(ip, endpoint) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded. Please slow down."})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthMiddleware validates user authentication
func AuthMiddleware(app *models.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		user, valid := app.AuthService.ValidateToken(token)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store user in context for handlers to use
		c.Set("user", user)
		c.Next()
	}
}

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware(app *models.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		userObj := user.(*models.User)
		if userObj.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken extracts the token from the Authorization header or query parameter
func extractToken(c *gin.Context) string {
	// Try Authorization header first (Bearer token)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Format: "Bearer <token>"
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:]
		}
		// Also support just the token without "Bearer " prefix
		return authHeader
	}

	// Try query parameter as fallback
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}
