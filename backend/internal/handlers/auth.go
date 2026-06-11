package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/backend/internal/models"
)

// Handler holds the application dependencies for handlers
type Handler struct {
	App *models.App
}

// login handles user login
func (h *Handler) Login(c *gin.Context) {
	if h.App.SettingsService != nil && !h.App.SettingsService.GetSettings().Auth.PasswordLoginEnabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Password login is disabled"})
		return
	}

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

// ListAuthProviders returns enabled third-party login providers.
func (h *Handler) ListAuthProviders(c *gin.Context) {
	passwordLoginEnabled := true
	if h.App.SettingsService != nil {
		passwordLoginEnabled = h.App.SettingsService.GetSettings().Auth.PasswordLoginEnabled
	}
	if h.App.OAuthService == nil {
		c.JSON(http.StatusOK, models.AuthProvidersResponse{
			Providers:            []models.AuthProviderResponse{},
			PasswordLoginEnabled: passwordLoginEnabled,
		})
		return
	}

	c.JSON(http.StatusOK, models.AuthProvidersResponse{
		Providers:            h.App.OAuthService.ListProviders(),
		PasswordLoginEnabled: passwordLoginEnabled,
	})
}

// StartOAuthLogin redirects the browser to a third-party authorization page.
func (h *Handler) StartOAuthLogin(c *gin.Context) {
	if h.App.OAuthService == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OAuth login is not enabled"})
		return
	}

	authURL, err := h.App.OAuthService.StartLogin(c.Request.Context(), c.Param("provider"))
	if err != nil {
		log.Printf("Failed to start OAuth login for provider '%s': %v", c.Param("provider"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// HandleOAuthCallback handles the provider callback and creates a short-lived login handoff.
func (h *Handler) HandleOAuthCallback(c *gin.Context) {
	if h.App.OAuthService == nil {
		c.Redirect(http.StatusFound, "/login.html?oauth=error")
		return
	}

	handoffCookie, err := h.App.OAuthService.HandleCallback(
		c.Request.Context(),
		c.Param("provider"),
		c.Query("code"),
		c.Query("state"),
	)
	if err != nil {
		log.Printf("OAuth callback failed for provider '%s': %v", c.Param("provider"), err)
		c.Redirect(http.StatusFound, "/login.html?oauth=error")
		return
	}

	http.SetCookie(c.Writer, handoffCookie)
	c.Redirect(http.StatusFound, "/login.html?oauth=complete")
}

// CompleteOAuthLogin exchanges the short-lived handoff for the existing login response shape.
func (h *Handler) CompleteOAuthLogin(c *gin.Context) {
	if h.App.OAuthService == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OAuth login is not enabled"})
		return
	}

	cookie, err := c.Request.Cookie(models.OAuthHandoffCookieName)
	if err != nil {
		_, clearCookie, _ := h.App.OAuthService.CompleteLogin("")
		http.SetCookie(c.Writer, clearCookie)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OAuth login handoff is missing"})
		return
	}

	response, clearCookie, err := h.App.OAuthService.CompleteLogin(cookie.Value)
	http.SetCookie(c.Writer, clearCookie)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
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
