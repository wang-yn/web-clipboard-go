package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/backend/internal/models"
)

type fakeHandlerOAuthService struct{}

func (fakeHandlerOAuthService) ListProviders() []models.AuthProviderResponse {
	return []models.AuthProviderResponse{{Name: "github", DisplayName: "GitHub"}}
}

func (fakeHandlerOAuthService) StartLogin(ctx context.Context, provider string) (string, error) {
	return "https://provider.example/auth?provider=" + provider, nil
}

func (fakeHandlerOAuthService) HandleCallback(ctx context.Context, provider, code, state string) (*http.Cookie, error) {
	return &http.Cookie{
		Name:     models.OAuthHandoffCookieName,
		Value:    "handoff-1",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func (fakeHandlerOAuthService) CompleteLogin(handoff string) (models.LoginResponse, *http.Cookie, error) {
	return models.LoginResponse{
			Token:     "local-token",
			User:      models.UserResponse{ID: "user-1", Username: "oauth-user", Role: "user", IsActive: true},
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}, &http.Cookie{
			Name:     models.OAuthHandoffCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}, nil
}

func TestListAuthProvidersReturnsConfiguredProviders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{App: &models.App{OAuthService: fakeHandlerOAuthService{}}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/auth/providers", nil)

	handler.ListAuthProviders(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Providers []models.AuthProviderResponse `json:"providers"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Providers) != 1 || response.Providers[0].Name != "github" {
		t.Fatalf("unexpected providers response: %#v", response.Providers)
	}
}

func TestStartOAuthLoginRedirectsToProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{App: &models.App{OAuthService: fakeHandlerOAuthService{}}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/auth/oauth/github/start", nil)
	context.Params = gin.Params{{Key: "provider", Value: "github"}}

	handler.StartOAuthLogin(context)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if location := recorder.Header().Get("Location"); location != "https://provider.example/auth?provider=github" {
		t.Fatalf("unexpected redirect location: %s", location)
	}
}

func TestCompleteOAuthLoginReturnsLoginResponseAndClearsHandoff(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{App: &models.App{OAuthService: fakeHandlerOAuthService{}}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/complete", nil)
	request.AddCookie(&http.Cookie{Name: models.OAuthHandoffCookieName, Value: "handoff-1"})
	context.Request = request

	handler.CompleteOAuthLogin(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response models.LoginResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Token != "local-token" || response.User.Username != "oauth-user" {
		t.Fatalf("unexpected login response: %#v", response)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != models.OAuthHandoffCookieName || cookies[0].MaxAge >= 0 {
		t.Fatalf("OAuth handoff cookie was not cleared: %#v", cookies)
	}
}
