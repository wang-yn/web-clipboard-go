package services

import (
	"context"
	"net/url"
	"testing"
)

type fakeOAuthProvider struct {
	name     string
	identity ExternalIdentity
}

func (p fakeOAuthProvider) Name() string {
	return p.name
}

func (p fakeOAuthProvider) DisplayName() string {
	return "Fake"
}

func (p fakeOAuthProvider) AuthCodeURL(state, codeChallenge, nonce string) string {
	values := url.Values{}
	values.Set("state", state)
	values.Set("challenge", codeChallenge)
	values.Set("nonce", nonce)
	return "https://provider.example/auth?" + values.Encode()
}

func (p fakeOAuthProvider) Exchange(ctx context.Context, code, verifier, nonce string) (*ExternalIdentity, error) {
	return &p.identity, nil
}

func TestOAuthCallbackAutoProvisionsUserAndCreatesLocalSession(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	authService := NewAuthService(manager)
	oauthService := NewOAuthService(manager, authService, OAuthSettings{
		BaseURL:       "https://clipboard.example.com",
		AutoProvision: true,
	}, []OAuthProvider{
		fakeOAuthProvider{
			name: "fake",
			identity: ExternalIdentity{
				Provider:      "fake",
				Subject:       "external-1",
				Email:         "new-user@example.com",
				EmailVerified: true,
				Username:      "new-user",
				DisplayName:   "New User",
			},
		},
	})

	authURL, err := oauthService.StartLogin(context.Background(), "fake")
	if err != nil {
		t.Fatal(err)
	}
	state := mustQueryParam(t, authURL, "state")

	handoffCookie, err := oauthService.HandleCallback(context.Background(), "fake", "code-1", state)
	if err != nil {
		t.Fatal(err)
	}
	login, clearCookie, err := oauthService.CompleteLogin(handoffCookie.Value)
	if err != nil {
		t.Fatal(err)
	}

	if clearCookie.MaxAge >= 0 {
		t.Fatalf("complete login must return a cookie deletion marker, got MaxAge=%d", clearCookie.MaxAge)
	}
	if login.Token == "" {
		t.Fatal("complete login did not return a local session token")
	}
	if _, ok := authService.ValidateToken(login.Token); !ok {
		t.Fatal("returned local session token is not valid")
	}
	if login.User.Role != "user" {
		t.Fatalf("auto-provisioned users must be ordinary users, got %q", login.User.Role)
	}

	linked := manager.GetUserByExternalIdentity("fake", "external-1")
	if linked == nil {
		t.Fatal("auto-provisioned user was not linked to external identity")
	}
}

func TestOAuthCallbackRejectsUnlinkedExistingVerifiedEmail(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateUser("local", "secret123", "same@example.com", "user"); err != nil {
		t.Fatal(err)
	}

	oauthService := NewOAuthService(manager, NewAuthService(manager), OAuthSettings{
		BaseURL:       "https://clipboard.example.com",
		AutoProvision: true,
	}, []OAuthProvider{
		fakeOAuthProvider{
			name: "fake",
			identity: ExternalIdentity{
				Provider:      "fake",
				Subject:       "external-2",
				Email:         "same@example.com",
				EmailVerified: true,
				Username:      "same",
			},
		},
	})

	authURL, err := oauthService.StartLogin(context.Background(), "fake")
	if err != nil {
		t.Fatal(err)
	}
	state := mustQueryParam(t, authURL, "state")

	if _, err := oauthService.HandleCallback(context.Background(), "fake", "code-1", state); err == nil {
		t.Fatal("expected verified email collision to be rejected")
	}
	if linked := manager.GetUserByExternalIdentity("fake", "external-2"); linked != nil {
		t.Fatalf("email collision should not create or link identity, got %#v", linked)
	}
}

func TestOAuthHandoffCanOnlyBeCompletedOnce(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	oauthService := NewOAuthService(manager, NewAuthService(manager), OAuthSettings{
		BaseURL:       "https://clipboard.example.com",
		AutoProvision: true,
	}, []OAuthProvider{
		fakeOAuthProvider{
			name: "fake",
			identity: ExternalIdentity{
				Provider:      "fake",
				Subject:       "external-3",
				Email:         "once@example.com",
				EmailVerified: true,
				Username:      "once",
			},
		},
	})

	authURL, err := oauthService.StartLogin(context.Background(), "fake")
	if err != nil {
		t.Fatal(err)
	}
	handoffCookie, err := oauthService.HandleCallback(context.Background(), "fake", "code-1", mustQueryParam(t, authURL, "state"))
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := oauthService.CompleteLogin(handoffCookie.Value); err != nil {
		t.Fatal(err)
	}
	if _, _, err := oauthService.CompleteLogin(handoffCookie.Value); err == nil {
		t.Fatal("expected reused OAuth handoff to be rejected")
	}
}

func TestOAuthServiceListsOnlyEnabledAndCompleteProvidersFromSettings(t *testing.T) {
	t.Setenv("APP_BASE_URL", "https://clipboard.example.com")

	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	settingsService, err := NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	settings := settingsService.GetSettings()
	settings.Auth.Google.Enabled = true
	settings.Auth.Google.ClientID = "google-client"
	settings.Auth.Google.ClientSecret = "google-secret"
	settings.Auth.GitHub.Enabled = true
	settings.Auth.GitHub.ClientID = "github-client"
	settings.Auth.GitHub.ClientSecret = ""
	if err := settingsService.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}

	oauthService := NewOAuthServiceFromSettings(manager, NewAuthService(manager), settingsService)
	providers := oauthService.ListProviders()
	if len(providers) != 1 || providers[0].Name != "google" {
		t.Fatalf("expected only complete google provider, got %#v", providers)
	}
}

func mustQueryParam(t *testing.T, rawURL, name string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	value := parsed.Query().Get(name)
	if value == "" {
		t.Fatalf("missing query parameter %q in %s", name, rawURL)
	}
	return value
}
