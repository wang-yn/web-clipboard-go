package handlers

import (
	cryptoRand "crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/backend/internal/models"
)

// SaveText handles saving text to clipboard
func (h *Handler) SaveText(c *gin.Context) {
	var request models.TextRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if !h.App.Security.ValidateContentRequest(c, request.Content) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request rejected for security reasons"})
		return
	}

	id := h.generateShortID()
	user := c.MustGet("user").(*models.User)
	createdAt := time.Now().UTC()
	item := &models.ClipboardItem{
		ID:        id,
		Type:      "text",
		UserID:    user.ID,
		Content:   request.Content,
		CreatedAt: createdAt,
		ExpiresAt: h.clipboardExpiresAt(createdAt),
	}

	h.App.DataMutex.Lock()
	h.App.ClipboardData[id] = item
	h.App.DataMutex.Unlock()

	c.JSON(http.StatusOK, models.SaveTextResponse{
		ID:        id,
		ExpiresAt: item.ExpiresAt,
	})
}

// GetText handles retrieving text from clipboard
func (h *Handler) GetText(c *gin.Context) {
	id := strings.ToLower(c.Param("id"))

	if !h.App.Security.ValidateAccessRequest(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied"})
		return
	}

	h.App.DataMutex.RLock()
	item, exists := h.App.ClipboardData[id]
	h.App.DataMutex.RUnlock()

	if !exists || item.Type != "text" || models.ClipboardItemExpired(item, time.Now().UTC()) {
		h.App.Security.LogAccess(c, id, "text", false)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found or expired"})
		return
	}

	h.App.Security.LogAccess(c, id, "text", true)
	c.JSON(http.StatusOK, models.GetTextResponse{
		Content:   item.Content,
		CreatedAt: item.CreatedAt,
	})
}

// SaveFile handles saving a file to clipboard
func (h *Handler) SaveFile(c *gin.Context) {
	if !h.App.Security.ValidateFileRequest(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request rejected for security reasons"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	if !h.App.Security.ValidateFileType(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	if header.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 50MB)"})
		return
	}

	id := h.generateShortID()
	user := c.MustGet("user").(*models.User)
	filePath := filepath.Join(h.App.TempDir, fmt.Sprintf("%s_%s", id, header.Filename))

	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	createdAt := time.Now().UTC()
	item := &models.ClipboardItem{
		ID:          id,
		Type:        "file",
		UserID:      user.ID,
		FileName:    header.Filename,
		FilePath:    filePath,
		ContentType: header.Header.Get("Content-Type"),
		CreatedAt:   createdAt,
		ExpiresAt:   h.clipboardExpiresAt(createdAt),
	}

	h.App.DataMutex.Lock()
	h.App.ClipboardData[id] = item
	h.App.DataMutex.Unlock()

	c.JSON(http.StatusOK, models.SaveFileResponse{
		ID:        id,
		FileName:  header.Filename,
		ExpiresAt: item.ExpiresAt,
	})
}

// ListRecentItems returns the current user's unexpired clipboard items.
func (h *Handler) ListRecentItems(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	now := time.Now().UTC()
	items := make([]models.RecentItemResponse, 0)

	h.App.DataMutex.RLock()
	for _, item := range h.App.ClipboardData {
		if item.UserID != user.ID || models.ClipboardItemExpired(item, now) {
			continue
		}
		items = append(items, toRecentItemResponse(item))
	}
	h.App.DataMutex.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if len(items) > 10 {
		items = items[:10]
	}

	c.JSON(http.StatusOK, models.ListRecentItemsResponse{Items: items})
}

func toRecentItemResponse(item *models.ClipboardItem) models.RecentItemResponse {
	description := item.FileName
	if item.Type == "text" {
		description = textDescription(item.Content)
	}
	return models.RecentItemResponse{
		ID:          item.ID,
		Type:        item.Type,
		Description: description,
		FileName:    item.FileName,
		CreatedAt:   item.CreatedAt,
		ExpiresAt:   item.ExpiresAt,
	}
}

func textDescription(content string) string {
	content = strings.TrimSpace(content)
	if len([]rune(content)) <= 50 {
		return content
	}
	return string([]rune(content)[:50]) + "..."
}

// GetFile handles retrieving a file from clipboard
func (h *Handler) GetFile(c *gin.Context) {
	id := strings.ToLower(c.Param("id"))

	if !h.App.Security.ValidateAccessRequest(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied"})
		return
	}

	h.App.DataMutex.RLock()
	item, exists := h.App.ClipboardData[id]
	h.App.DataMutex.RUnlock()

	if !exists || item.Type != "file" || models.ClipboardItemExpired(item, time.Now().UTC()) {
		h.App.Security.LogAccess(c, id, "file", false)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found or expired"})
		return
	}

	if _, err := os.Stat(item.FilePath); os.IsNotExist(err) {
		h.App.Security.LogAccess(c, id, "file", false)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}

	h.App.Security.LogAccess(c, id, "file", true)
	c.Header("Content-Disposition", contentDispositionHeader(item.FileName))
	c.File(item.FilePath)
}

func contentDispositionHeader(fileName string) string {
	return fmt.Sprintf(
		"attachment; filename=\"%s\"; filename*=UTF-8''%s",
		asciiFallbackFileName(fileName),
		url.PathEscape(fileName),
	)
}

func asciiFallbackFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	if !isSafeASCIIFileName(fileName) {
		if isSafeASCIIFileName(ext) {
			return "download" + ext
		}
		return "download"
	}
	return fileName
}

func isSafeASCIIFileName(fileName string) bool {
	if fileName == "" {
		return false
	}
	for _, r := range fileName {
		if r < 0x20 || r > 0x7e || r == '"' || r == '\\' || r == ';' {
			return false
		}
	}
	return true
}

// DeleteItem handles deleting a clipboard item
func (h *Handler) DeleteItem(c *gin.Context) {
	id := strings.ToLower(c.Param("id"))

	h.App.DataMutex.Lock()
	item, exists := h.App.ClipboardData[id]
	if exists {
		delete(h.App.ClipboardData, id)
	}
	h.App.DataMutex.Unlock()

	if exists && item.Type == "file" && item.FilePath != "" {
		os.Remove(item.FilePath)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

// Cleanup handles cleaning up expired items
func (h *Handler) Cleanup(c *gin.Context) {
	removedCount := 0
	now := time.Now().UTC()

	h.App.DataMutex.Lock()
	for id, item := range h.App.ClipboardData {
		if models.ClipboardItemExpired(item, now) {
			if item.Type == "file" && item.FilePath != "" {
				os.Remove(item.FilePath)
			}
			delete(h.App.ClipboardData, id)
			removedCount++
		}
	}
	h.App.DataMutex.Unlock()

	c.JSON(http.StatusOK, models.CleanupResponse{
		RemovedCount: removedCount,
	})
}

// generateShortID generates a unique short ID for clipboard items
func (h *Handler) generateShortID() string {
	h.App.DataMutex.RLock()
	defer h.App.DataMutex.RUnlock()

	for attempt := 0; attempt < 100; attempt++ {
		id, err := generateRandomString(4)
		if err != nil {
			continue
		}
		if _, exists := h.App.ClipboardData[id]; !exists {
			return id
		}
	}

	id, _ := generateRandomString(6)
	return id
}

func (h *Handler) clipboardExpiresAt(now time.Time) time.Time {
	settings := models.DefaultSystemSettings()
	if h.App.SettingsService != nil {
		settings = h.App.SettingsService.GetSettings()
	}
	return settings.Clipboard.ExpiresAt(now)
}

// generateRandomString generates a random alphanumeric string
func generateRandomString(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		num, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(chars[num.Int64()])
	}
	return sb.String(), nil
}
