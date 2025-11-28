package handlers

import (
	cryptoRand "crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/internal/models"
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
	item := &models.ClipboardItem{
		ID:        id,
		Type:      "text",
		Content:   request.Content,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
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

	if !exists || item.Type != "text" || item.ExpiresAt.Before(time.Now().UTC()) {
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

	item := &models.ClipboardItem{
		ID:          id,
		Type:        "file",
		FileName:    header.Filename,
		FilePath:    filePath,
		ContentType: header.Header.Get("Content-Type"),
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(10 * time.Minute),
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

	if !exists || item.Type != "file" || item.ExpiresAt.Before(time.Now().UTC()) {
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
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", item.FileName))
	c.File(item.FilePath)
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
		if item.ExpiresAt.Before(now) {
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
