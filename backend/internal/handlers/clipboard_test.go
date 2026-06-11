package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/backend/internal/models"
	"web-clipboard-go/backend/internal/services"
)

type allowSecurityService struct{}

func (allowSecurityService) ValidateContentRequest(c interface{}, content string) bool  { return true }
func (allowSecurityService) ValidateFileRequest(c interface{}) bool                     { return true }
func (allowSecurityService) ValidateFileType(fileName string) bool                      { return true }
func (allowSecurityService) ValidateAccessRequest(c interface{}) bool                   { return true }
func (allowSecurityService) LogAccess(c interface{}, id, itemType string, success bool) {}
func (allowSecurityService) CleanupExpired()                                            {}
func (allowSecurityService) GetClientIP(c interface{}) string                           { return "127.0.0.1" }

func TestGetFileUsesRFC5987FilenameForUnicodeDownloads(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "stored-file")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	app := &models.App{
		ClipboardData: map[string]*models.ClipboardItem{
			"abc1": {
				ID:        "abc1",
				Type:      "file",
				FileName:  "中文 报告.txt",
				FilePath:  filePath,
				ExpiresAt: time.Now().UTC().Add(time.Minute),
			},
		},
		DataMutex: &sync.RWMutex{},
		Security:  allowSecurityService{},
	}
	handler := &Handler{App: app}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("GET", "/api/file/abc1", nil)
	context.Params = gin.Params{{Key: "id", Value: "abc1"}}

	handler.GetFile(context)

	disposition := recorder.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, `filename="download.txt"`) {
		t.Fatalf("expected ASCII fallback filename, got %q", disposition)
	}
	if !strings.Contains(disposition, `filename*=UTF-8''%E4%B8%AD%E6%96%87%20%E6%8A%A5%E5%91%8A.txt`) {
		t.Fatalf("expected RFC 5987 UTF-8 filename, got %q", disposition)
	}
}

func TestListRecentItemsShowsCurrentUsersUnexpiredItemsAcrossSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	app := &models.App{
		ClipboardData: map[string]*models.ClipboardItem{
			"same1": {
				ID:        "same1",
				Type:      "text",
				UserID:    "user-1",
				Content:   "text saved from another browser",
				CreatedAt: now.Add(-1 * time.Minute),
				ExpiresAt: now.Add(9 * time.Minute),
			},
			"same2": {
				ID:        "same2",
				Type:      "file",
				UserID:    "user-1",
				FileName:  "notes.txt",
				CreatedAt: now.Add(-2 * time.Minute),
				ExpiresAt: now.Add(8 * time.Minute),
			},
			"other": {
				ID:        "other",
				Type:      "text",
				UserID:    "user-2",
				Content:   "other user's text",
				CreatedAt: now,
				ExpiresAt: now.Add(10 * time.Minute),
			},
			"expired": {
				ID:        "expired",
				Type:      "text",
				UserID:    "user-1",
				Content:   "expired text",
				CreatedAt: now.Add(-11 * time.Minute),
				ExpiresAt: now.Add(-1 * time.Minute),
			},
		},
		DataMutex: &sync.RWMutex{},
		Security:  allowSecurityService{},
	}
	handler := &Handler{App: app}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("GET", "/api/items", nil)
	context.Set("user", &models.User{ID: "user-1", Username: "same-user"})

	handler.ListRecentItems(context)

	if recorder.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response models.ListRecentItemsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Items) != 2 {
		t.Fatalf("expected 2 current-user items, got %d: %#v", len(response.Items), response.Items)
	}
	if response.Items[0].ID != "same1" || response.Items[1].ID != "same2" {
		t.Fatalf("expected current-user items newest first, got %#v", response.Items)
	}
	for _, item := range response.Items {
		if item.ID == "other" || item.ID == "expired" {
			t.Fatalf("unexpected item in current-user recent list: %#v", item)
		}
	}
}

func TestSaveTextUsesConfiguredClipboardExpiration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settingsService, err := services.NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	settings := settingsService.GetSettings()
	settings.Clipboard.ExpirationValue = 2
	settings.Clipboard.ExpirationUnit = models.ClipboardExpirationUnitHour
	if err := settingsService.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}
	app := &models.App{
		ClipboardData:   map[string]*models.ClipboardItem{},
		DataMutex:       &sync.RWMutex{},
		Security:        allowSecurityService{},
		SettingsService: settingsService,
	}
	handler := &Handler{App: app}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("POST", "/api/text", strings.NewReader(`{"content":"hello"}`))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("user", &models.User{ID: "user-1", Username: "same-user"})

	before := time.Now().UTC()
	handler.SaveText(context)
	after := time.Now().UTC()

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	for _, item := range app.ClipboardData {
		min := before.Add(2 * time.Hour)
		max := after.Add(2 * time.Hour)
		if item.ExpiresAt.Before(min) || item.ExpiresAt.After(max) {
			t.Fatalf("item expiration %v outside configured range %v..%v", item.ExpiresAt, min, max)
		}
		return
	}
	t.Fatal("saved text item missing")
}

func TestCleanupKeepsNeverExpiringItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	app := &models.App{
		ClipboardData: map[string]*models.ClipboardItem{
			"never": {
				ID:        "never",
				Type:      "text",
				UserID:    "user-1",
				Content:   "keep",
				ExpiresAt: time.Time{},
			},
			"expired": {
				ID:        "expired",
				Type:      "text",
				UserID:    "user-1",
				Content:   "remove",
				ExpiresAt: now.Add(-time.Minute),
			},
		},
		DataMutex: &sync.RWMutex{},
		Security:  allowSecurityService{},
	}
	handler := &Handler{App: app}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("GET", "/api/cleanup", nil)

	handler.Cleanup(context)

	if _, exists := app.ClipboardData["never"]; !exists {
		t.Fatal("never-expiring item should remain after cleanup")
	}
	if _, exists := app.ClipboardData["expired"]; exists {
		t.Fatal("expired item should be removed by cleanup")
	}
}
