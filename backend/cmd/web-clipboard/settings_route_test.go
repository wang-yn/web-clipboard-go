package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"web-clipboard-go/backend/internal/models"
	"web-clipboard-go/backend/internal/services"
)

func TestSettingsRouteRequiresAdminUser(t *testing.T) {
	userManager, err := services.NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	user, err := userManager.CreateUser("normal", "password123", "normal@example.com", "user")
	if err != nil {
		t.Fatal(err)
	}
	authService := services.NewAuthService(userManager)
	session, err := authService.CreateSession(user.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	settingsService, err := services.NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	router := setupRouter(&models.App{
		ClipboardData:   map[string]*models.ClipboardItem{},
		DataMutex:       &sync.RWMutex{},
		RateLimiter:     services.NewRateLimitService(),
		Security:        services.NewSecurityService(),
		UserManager:     userManager,
		AuthService:     authService,
		SettingsService: settingsService,
	})
	request := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	request.Header.Set("Authorization", "Bearer "+session.Token)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected non-admin settings request to be forbidden, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
