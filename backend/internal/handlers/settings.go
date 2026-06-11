package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/backend/internal/models"
)

// GetSettings returns sanitized system settings for administrators.
func (h *Handler) GetSettings(c *gin.Context) {
	if h.App.SettingsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Settings service is not available"})
		return
	}

	c.JSON(http.StatusOK, h.App.SettingsService.GetSettingsResponse())
}

// UpdateSettings updates system settings.
func (h *Handler) UpdateSettings(c *gin.Context) {
	if h.App.SettingsService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Settings service is not available"})
		return
	}

	var req models.SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if err := h.App.SettingsService.SaveSettings(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.App.SettingsService.GetSettingsResponse())
}
