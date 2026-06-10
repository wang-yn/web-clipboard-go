package handlers

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/internal/models"
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
