package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"web-clipboard-go/backend/internal/models"
)

const (
	oauthStateTTL   = 10 * time.Minute
	oauthHandoffTTL = time.Minute
)

type OAuthSettings struct {
	BaseURL             string
	AutoProvision       bool
	AllowedEmailDomains []string
}

type OAuthProvider interface {
	Name() string
	DisplayName() string
	AuthCodeURL(state, codeChallenge, nonce string) string
	Exchange(ctx context.Context, code, verifier, nonce string) (*ExternalIdentity, error)
}

type OAuthService struct {
	userManager *UserManager
	authService *AuthService
	settings    OAuthSettings
	providers   map[string]OAuthProvider
	states      map[string]oauthState
	handoffs    map[string]oauthHandoff
	mutex       sync.Mutex
}

type oauthState struct {
	Provider  string
	Verifier  string
	Nonce     string
	ExpiresAt time.Time
}

type oauthHandoff struct {
	Response  models.LoginResponse
	ExpiresAt time.Time
}

type GoogleProvider struct {
	name        string
	displayName string
	config      oauth2.Config
	issuer      string
	clientID    string
}

type GitHubProvider struct {
	name        string
	displayName string
	config      oauth2.Config
	apiBaseURL  string
}

func NewOAuthService(userManager *UserManager, authService *AuthService, settings OAuthSettings, providers []OAuthProvider) *OAuthService {
	providerMap := make(map[string]OAuthProvider, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(provider.Name()))
		if name != "" {
			providerMap[name] = provider
		}
	}

	return &OAuthService{
		userManager: userManager,
		authService: authService,
		settings:    normalizeOAuthSettings(settings),
		providers:   providerMap,
		states:      make(map[string]oauthState),
		handoffs:    make(map[string]oauthHandoff),
	}
}

func NewOAuthServiceFromEnv(userManager *UserManager, authService *AuthService) *OAuthService {
	settings := OAuthSettings{
		BaseURL:             strings.TrimSpace(os.Getenv("APP_BASE_URL")),
		AutoProvision:       parseBoolEnv("OAUTH_AUTO_PROVISION", false),
		AllowedEmailDomains: splitCSV(os.Getenv("OAUTH_ALLOWED_EMAIL_DOMAINS")),
	}

	var providers []OAuthProvider
	if parseBoolEnv("GOOGLE_OAUTH_ENABLED", false) {
		if provider := NewGoogleProviderFromEnv(settings.BaseURL); provider != nil {
			providers = append(providers, provider)
		}
	}
	if parseBoolEnv("GITHUB_OAUTH_ENABLED", false) {
		if provider := NewGitHubProviderFromEnv(settings.BaseURL); provider != nil {
			providers = append(providers, provider)
		}
	}

	return NewOAuthService(userManager, authService, settings, providers)
}

func NewGoogleProviderFromEnv(baseURL string) OAuthProvider {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))
	if clientID == "" || clientSecret == "" || strings.TrimSpace(baseURL) == "" {
		return nil
	}

	return &GoogleProvider{
		name:        "google",
		displayName: "Google",
		clientID:    clientID,
		issuer:      "https://accounts.google.com",
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			RedirectURL: buildOAuthRedirectURL(baseURL, "google"),
			Scopes:      []string{"openid", "email", "profile"},
		},
	}
}

func NewGitHubProviderFromEnv(baseURL string) OAuthProvider {
	clientID := strings.TrimSpace(os.Getenv("GITHUB_OAUTH_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"))
	if clientID == "" || clientSecret == "" || strings.TrimSpace(baseURL) == "" {
		return nil
	}

	return &GitHubProvider{
		name:        "github",
		displayName: "GitHub",
		apiBaseURL:  "https://api.github.com",
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
			RedirectURL: buildOAuthRedirectURL(baseURL, "github"),
			Scopes:      []string{"read:user", "user:email"},
		},
	}
}

func (s *OAuthService) ListProviders() []models.AuthProviderResponse {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cleanupLocked(time.Now().UTC())

	providers := make([]models.AuthProviderResponse, 0, len(s.providers))
	for _, provider := range s.providers {
		providers = append(providers, models.AuthProviderResponse{
			Name:        provider.Name(),
			DisplayName: provider.DisplayName(),
		})
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})
	return providers
}

func (s *OAuthService) StartLogin(ctx context.Context, providerName string) (string, error) {
	provider, err := s.getProvider(providerName)
	if err != nil {
		return "", err
	}

	state, err := randomURLToken(32)
	if err != nil {
		return "", err
	}
	verifier, err := randomURLToken(64)
	if err != nil {
		return "", err
	}
	nonce, err := randomURLToken(32)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	s.mutex.Lock()
	s.cleanupLocked(now)
	s.states[state] = oauthState{
		Provider:  provider.Name(),
		Verifier:  verifier,
		Nonce:     nonce,
		ExpiresAt: now.Add(oauthStateTTL),
	}
	s.mutex.Unlock()

	return provider.AuthCodeURL(state, pkceChallenge(verifier), nonce), nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, providerName, code, stateValue string) (*http.Cookie, error) {
	provider, err := s.getProvider(providerName)
	if err != nil {
		return nil, err
	}
	code = strings.TrimSpace(code)
	stateValue = strings.TrimSpace(stateValue)
	if code == "" {
		return nil, errors.New("oauth code is required")
	}
	if stateValue == "" {
		return nil, errors.New("oauth state is required")
	}

	now := time.Now().UTC()
	s.mutex.Lock()
	s.cleanupLocked(now)
	state, exists := s.states[stateValue]
	if exists {
		delete(s.states, stateValue)
	}
	s.mutex.Unlock()
	if !exists {
		return nil, errors.New("invalid or expired oauth state")
	}
	if state.ExpiresAt.Before(now) {
		return nil, errors.New("invalid or expired oauth state")
	}
	if state.Provider != provider.Name() {
		return nil, errors.New("oauth provider mismatch")
	}

	identity, err := provider.Exchange(ctx, code, state.Verifier, state.Nonce)
	if err != nil {
		return nil, err
	}
	localUser, err := s.resolveLocalUser(*identity)
	if err != nil {
		return nil, err
	}

	session, err := s.authService.CreateSession(localUser.ID, false)
	if err != nil {
		return nil, err
	}
	handoff, err := randomURLToken(32)
	if err != nil {
		return nil, err
	}
	response := models.LoginResponse{
		Token:     session.Token,
		User:      models.ToUserResponse(localUser),
		ExpiresAt: session.ExpiresAt,
	}

	s.mutex.Lock()
	s.cleanupLocked(now)
	s.handoffs[handoff] = oauthHandoff{
		Response:  response,
		ExpiresAt: now.Add(oauthHandoffTTL),
	}
	s.mutex.Unlock()

	return &http.Cookie{
		Name:     models.OAuthHandoffCookieName,
		Value:    handoff,
		Path:     "/",
		MaxAge:   int(oauthHandoffTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(s.settings.BaseURL), "https://"),
	}, nil
}

func (s *OAuthService) CompleteLogin(handoff string) (models.LoginResponse, *http.Cookie, error) {
	handoff = strings.TrimSpace(handoff)
	if handoff == "" {
		return models.LoginResponse{}, clearOAuthHandoffCookie(s.settings), errors.New("oauth handoff is required")
	}

	now := time.Now().UTC()
	s.mutex.Lock()
	s.cleanupLocked(now)
	result, exists := s.handoffs[handoff]
	if exists {
		delete(s.handoffs, handoff)
	}
	s.mutex.Unlock()
	if !exists || result.ExpiresAt.Before(now) {
		return models.LoginResponse{}, clearOAuthHandoffCookie(s.settings), errors.New("invalid or expired oauth handoff")
	}

	return result.Response, clearOAuthHandoffCookie(s.settings), nil
}

func (s *OAuthService) resolveLocalUser(identity ExternalIdentity) (*models.User, error) {
	identity = normalizeExternalIdentity(identity)
	if err := validateExternalIdentity(identity); err != nil {
		return nil, err
	}
	if !s.emailDomainAllowed(identity.Email) {
		return nil, errors.New("email domain is not allowed")
	}

	if user := s.userManager.GetUserByExternalIdentity(identity.Provider, identity.Subject); user != nil {
		if !user.IsActive {
			return nil, errors.New("user account is disabled")
		}
		return user, nil
	}

	if identity.EmailVerified {
		if existing := s.userManager.GetUserByVerifiedEmail(identity.Email); existing != nil {
			return nil, errors.New("verified email already belongs to a local account")
		}
	}
	if !s.settings.AutoProvision {
		return nil, errors.New("oauth account is not linked")
	}
	return s.userManager.CreateExternalUser(identity, "user")
}

func (s *OAuthService) getProvider(providerName string) (OAuthProvider, error) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName == "" {
		return nil, errors.New("oauth provider is required")
	}

	s.mutex.Lock()
	provider := s.providers[providerName]
	s.mutex.Unlock()
	if provider == nil {
		return nil, errors.New("oauth provider is not enabled")
	}
	return provider, nil
}

func (s *OAuthService) emailDomainAllowed(email string) bool {
	if len(s.settings.AllowedEmailDomains) == 0 {
		return true
	}
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	if len(parts) != 2 || parts[1] == "" {
		return false
	}
	domain := parts[1]
	for _, allowed := range s.settings.AllowedEmailDomains {
		if domain == allowed {
			return true
		}
	}
	return false
}

func (s *OAuthService) cleanupLocked(now time.Time) {
	for key, state := range s.states {
		if state.ExpiresAt.Before(now) {
			delete(s.states, key)
		}
	}
	for key, handoff := range s.handoffs {
		if handoff.ExpiresAt.Before(now) {
			delete(s.handoffs, key)
		}
	}
}

func (p *GoogleProvider) Name() string {
	return p.name
}

func (p *GoogleProvider) DisplayName() string {
	return p.displayName
}

func (p *GoogleProvider) AuthCodeURL(state, codeChallenge, nonce string) string {
	return p.config.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code, verifier, nonce string) (*ExternalIdentity, error) {
	token, err := p.config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange google oauth code: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("google id_token is missing")
	}

	provider, err := oidc.NewProvider(ctx, p.issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize google oidc provider: %w", err)
	}
	idToken, err := provider.Verifier(&oidc.Config{ClientID: p.clientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify google id_token: %w", err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Nonce         string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse google id_token claims: %w", err)
	}
	if claims.Nonce != nonce {
		return nil, errors.New("google id_token nonce mismatch")
	}
	if !claims.EmailVerified {
		return nil, errors.New("google email must be verified")
	}

	return &ExternalIdentity{
		Provider:      p.name,
		Subject:       idToken.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Username:      strings.Split(claims.Email, "@")[0],
		DisplayName:   claims.Name,
		AvatarURL:     claims.Picture,
	}, nil
}

func (p *GitHubProvider) Name() string {
	return p.name
}

func (p *GitHubProvider) DisplayName() string {
	return p.displayName
}

func (p *GitHubProvider) AuthCodeURL(state, codeChallenge, nonce string) string {
	return p.config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

func (p *GitHubProvider) Exchange(ctx context.Context, code, verifier, nonce string) (*ExternalIdentity, error) {
	token, err := p.config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange github oauth code: %w", err)
	}
	client := p.config.Client(ctx, token)

	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := githubGetJSON(ctx, client, p.apiBaseURL+"/user", &user); err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, errors.New("github user id is missing")
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := githubGetJSON(ctx, client, p.apiBaseURL+"/user/emails", &emails); err != nil {
		return nil, err
	}
	email := ""
	for _, item := range emails {
		if item.Primary && item.Verified {
			email = item.Email
			break
		}
	}
	if email == "" {
		return nil, errors.New("github account must have a verified primary email")
	}

	displayName := user.Name
	if displayName == "" {
		displayName = user.Login
	}
	return &ExternalIdentity{
		Provider:      p.name,
		Subject:       fmt.Sprintf("%d", user.ID),
		Email:         email,
		EmailVerified: true,
		Username:      user.Login,
		DisplayName:   displayName,
		AvatarURL:     user.AvatarURL,
	}, nil
}

func githubGetJSON(ctx context.Context, client *http.Client, endpoint string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("github api request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func normalizeOAuthSettings(settings OAuthSettings) OAuthSettings {
	settings.BaseURL = strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")
	settings.AllowedEmailDomains = normalizeDomains(settings.AllowedEmailDomains)
	return settings
}

func normalizeDomains(domains []string) []string {
	seen := make(map[string]bool, len(domains))
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" || seen[domain] {
			continue
		}
		seen[domain] = true
		result = append(result, domain)
	}
	sort.Strings(result)
	return result
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func parseBoolEnv(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func buildOAuthRedirectURL(baseURL, provider string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return baseURL + "/api/auth/oauth/" + url.PathEscape(provider) + "/callback"
}

func randomURLToken(length int) (string, error) {
	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func clearOAuthHandoffCookie(settings OAuthSettings) *http.Cookie {
	return &http.Cookie{
		Name:     models.OAuthHandoffCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(strings.TrimSpace(settings.BaseURL)), "https://"),
	}
}
