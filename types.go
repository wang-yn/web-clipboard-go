package main

import (
	"sync"
	"time"
)

type App struct {
	clipboardData map[string]*ClipboardItem
	dataMutex     sync.RWMutex
	tempDir       string
	rateLimiter   *RateLimitService
	security      *SecurityService
	cleanupTicker *time.Ticker
}

type ClipboardItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Content     string    `json:"content,omitempty"`
	FileName    string    `json:"fileName,omitempty"`
	FilePath    string    `json:"-"`
	ContentType string    `json:"contentType,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type TextRequest struct {
	Content string `json:"content" binding:"required"`
}

type SaveTextResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type GetTextResponse struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type SaveFileResponse struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type CleanupResponse struct {
	RemovedCount int `json:"removedCount"`
}
