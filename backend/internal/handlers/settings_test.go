package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/backend/internal/models"
	"web-clipboard-go/backend/internal/services"
)

func TestGetSettingsDoesNotReturnSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settingsService, err := services.NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	settings := settingsService.GetSettings()
	settings.Auth.Google.Enabled = true
	settings.Auth.Google.ClientID = "google-client"
	settings.Auth.Google.ClientSecret = "google-secret"
	if err := settingsService.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}
	handler := &Handler{App: &models.App{SettingsService: settingsService}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/settings", nil)

	handler.GetSettings(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response models.SystemSettingsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Auth.Google.ClientSecret != "" {
		t.Fatal("settings response leaked google client secret")
	}
	if !response.Auth.Google.ClientSecretSet {
		t.Fatal("settings response did not report configured google secret")
	}
}

func TestUpdateSettingsPreservesExistingSecretWhenBlank(t *testing.T) {
	gin.SetMode(gin.TestMode)
	settingsService, err := services.NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	settings := settingsService.GetSettings()
	settings.Auth.Google.Enabled = true
	settings.Auth.Google.ClientID = "google-client"
	settings.Auth.Google.ClientSecret = "old-secret"
	if err := settingsService.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}
	handler := &Handler{App: &models.App{SettingsService: settingsService}}
	body := `{
		"auth": {
			"passwordLoginEnabled": true,
			"oauthAutoProvision": false,
			"allowedEmailDomains": [],
			"google": { "enabled": true, "clientId": "new-google-client", "clientSecret": "" },
			"github": { "enabled": false, "clientId": "", "clientSecret": "" }
		},
		"clipboard": { "expirationValue": 2, "expirationUnit": "hour" }
	}`
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewBufferString(body))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	updated := settingsService.GetSettings()
	if updated.Auth.Google.ClientSecret != "old-secret" {
		t.Fatalf("blank secret should preserve old value, got %q", updated.Auth.Google.ClientSecret)
	}
	if updated.Auth.Google.ClientID != "new-google-client" {
		t.Fatalf("client id was not updated: %q", updated.Auth.Google.ClientID)
	}
	if updated.Clipboard.ExpirationValue != 2 || updated.Clipboard.ExpirationUnit != models.ClipboardExpirationUnitHour {
		t.Fatalf("clipboard settings were not updated: %#v", updated.Clipboard)
	}
}
