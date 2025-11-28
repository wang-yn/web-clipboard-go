package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/internal/models"
)

// Handler holds the application dependencies for handlers
type Handler struct {
	App *models.App
}

// login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate credentials
	user, err := h.App.UserManager.ValidateCredentials(req.Username, req.Password)
	if err != nil {
		log.Printf("Login failed for user '%s': %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Create session
	session, err := h.App.AuthService.CreateSession(user.ID, req.RememberMe)
	if err != nil {
		log.Printf("Failed to create session for user '%s': %v", req.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	log.Printf("User '%s' logged in successfully (RememberMe: %v)", user.Username, req.RememberMe)

	// Return login response
	c.JSON(http.StatusOK, models.LoginResponse{
		Token:     session.Token,
		User:      models.ToUserResponse(user),
		ExpiresAt: session.ExpiresAt,
	})
}

// logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token provided"})
		return
	}

	// Get user info before deleting session for logging
	user, _ := h.App.AuthService.GetUserByToken(token)

	// Delete session
	h.App.AuthService.DeleteSession(token)

	if user != nil {
		log.Printf("User '%s' logged out", user.Username)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// getCurrentUser returns the current logged-in user's information
func (h *Handler) GetCurrentUser(c *gin.Context) {
	// User is already validated by authMiddleware and stored in context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)
	c.JSON(http.StatusOK, models.ToUserResponse(userObj))
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
