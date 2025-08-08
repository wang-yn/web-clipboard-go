package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (app *App) saveText(c *gin.Context) {
	if app.security == nil {
		app.security = NewSecurityService()
	}

	var request TextRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if !app.security.ValidateContentRequest(c, request.Content) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request rejected for security reasons"})
		return
	}

	id := app.generateShortID()
	item := &ClipboardItem{
		ID:        id,
		Type:      "text",
		Content:   request.Content,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
	}

	app.dataMutex.Lock()
	app.clipboardData[id] = item
	app.dataMutex.Unlock()

	c.JSON(http.StatusOK, SaveTextResponse{
		ID:        id,
		ExpiresAt: item.ExpiresAt,
	})
}

func (app *App) getText(c *gin.Context) {
	if app.security == nil {
		app.security = NewSecurityService()
	}

	id := strings.ToLower(c.Param("id"))

	if !app.security.ValidateAccessRequest(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied"})
		return
	}

	app.dataMutex.RLock()
	item, exists := app.clipboardData[id]
	app.dataMutex.RUnlock()

	if !exists || item.Type != "text" || item.ExpiresAt.Before(time.Now().UTC()) {
		app.security.LogAccess(c, id, "text", false)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found or expired"})
		return
	}

	app.security.LogAccess(c, id, "text", true)
	c.JSON(http.StatusOK, GetTextResponse{
		Content:   item.Content,
		CreatedAt: item.CreatedAt,
	})
}

func (app *App) saveFile(c *gin.Context) {
	if app.security == nil {
		app.security = NewSecurityService()
	}

	if !app.security.ValidateFileRequest(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request rejected for security reasons"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	if !app.security.ValidateFileType(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	if header.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 50MB)"})
		return
	}

	id := app.generateShortID()
	filePath := filepath.Join(app.tempDir, fmt.Sprintf("%s_%s", id, header.Filename))

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

	item := &ClipboardItem{
		ID:          id,
		Type:        "file",
		FileName:    header.Filename,
		FilePath:    filePath,
		ContentType: header.Header.Get("Content-Type"),
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(10 * time.Minute),
	}

	app.dataMutex.Lock()
	app.clipboardData[id] = item
	app.dataMutex.Unlock()

	c.JSON(http.StatusOK, SaveFileResponse{
		ID:        id,
		FileName:  header.Filename,
		ExpiresAt: item.ExpiresAt,
	})
}

func (app *App) getFile(c *gin.Context) {
	if app.security == nil {
		app.security = NewSecurityService()
	}

	id := strings.ToLower(c.Param("id"))

	if !app.security.ValidateAccessRequest(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Access denied"})
		return
	}

	app.dataMutex.RLock()
	item, exists := app.clipboardData[id]
	app.dataMutex.RUnlock()

	if !exists || item.Type != "file" || item.ExpiresAt.Before(time.Now().UTC()) {
		app.security.LogAccess(c, id, "file", false)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found or expired"})
		return
	}

	if _, err := os.Stat(item.FilePath); os.IsNotExist(err) {
		app.security.LogAccess(c, id, "file", false)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}

	app.security.LogAccess(c, id, "file", true)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", item.FileName))
	c.File(item.FilePath)
}

func (app *App) deleteItem(c *gin.Context) {
	id := strings.ToLower(c.Param("id"))

	app.dataMutex.Lock()
	item, exists := app.clipboardData[id]
	if exists {
		delete(app.clipboardData, id)
	}
	app.dataMutex.Unlock()

	if exists && item.Type == "file" && item.FilePath != "" {
		os.Remove(item.FilePath)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

func (app *App) cleanup(c *gin.Context) {
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

	c.JSON(http.StatusOK, CleanupResponse{
		RemovedCount: removedCount,
	})
}