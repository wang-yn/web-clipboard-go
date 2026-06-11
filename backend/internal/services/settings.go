package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"web-clipboard-go/backend/internal/models"
)

type SettingsService struct {
	settings models.SystemSettings
	filePath string
	mutex    sync.RWMutex
}

func NewSettingsService(dataDir string) (*SettingsService, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	service := &SettingsService{
		settings: models.DefaultSystemSettings(),
		filePath: filepath.Join(dataDir, "settings.json"),
	}
	if err := service.loadSettings(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *SettingsService) GetSettings() models.SystemSettings {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return cloneSettings(s.settings)
}

func (s *SettingsService) GetSettingsResponse() models.SystemSettingsResponse {
	settings := s.GetSettings()
	settings.Auth.Google = sanitizeProviderConfig(settings.Auth.Google)
	settings.Auth.GitHub = sanitizeProviderConfig(settings.Auth.GitHub)
	return settings
}

func (s *SettingsService) SaveSettings(settings models.SystemSettings) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	current := cloneSettings(s.settings)
	next := normalizeSettingsForSave(settings, current)
	if err := validateSettings(next); err != nil {
		return err
	}
	if err := s.writeSettings(next); err != nil {
		return err
	}
	s.settings = next
	return nil
}

func (s *SettingsService) loadSettings() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.writeSettings(s.settings)
		}
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	settings := models.DefaultSystemSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}
	settings = normalizeLoadedSettings(settings)
	if err := validateSettings(settings); err != nil {
		return fmt.Errorf("invalid settings file: %w", err)
	}

	s.mutex.Lock()
	s.settings = settings
	s.mutex.Unlock()
	return nil
}

func (s *SettingsService) writeSettings(settings models.SystemSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	return nil
}

func normalizeLoadedSettings(settings models.SystemSettings) models.SystemSettings {
	defaults := models.DefaultSystemSettings()
	if settings.Clipboard.ExpirationUnit == "" {
		settings.Clipboard = defaults.Clipboard
	}
	if settings.Clipboard.ExpirationValue <= 0 && settings.Clipboard.ExpirationUnit != models.ClipboardExpirationUnitNever {
		settings.Clipboard.ExpirationValue = defaults.Clipboard.ExpirationValue
	}
	if settings.Auth.AllowedEmailDomains == nil {
		settings.Auth.AllowedEmailDomains = []string{}
	}
	settings.Auth.AllowedEmailDomains = normalizeEmailDomains(settings.Auth.AllowedEmailDomains)
	return settings
}

func normalizeSettingsForSave(settings, current models.SystemSettings) models.SystemSettings {
	settings = normalizeLoadedSettings(settings)
	settings.Auth.Google = normalizeProviderForSave(settings.Auth.Google, current.Auth.Google)
	settings.Auth.GitHub = normalizeProviderForSave(settings.Auth.GitHub, current.Auth.GitHub)
	if settings.Clipboard.ExpirationUnit == models.ClipboardExpirationUnitNever {
		settings.Clipboard.ExpirationValue = 0
	}
	return settings
}

func normalizeProviderForSave(next, current models.OAuthProviderConfig) models.OAuthProviderConfig {
	next.ClientID = strings.TrimSpace(next.ClientID)
	next.ClientSecret = strings.TrimSpace(next.ClientSecret)
	if next.ClearClientSecret {
		next.ClientSecret = ""
	} else if next.ClientSecret == "" {
		next.ClientSecret = current.ClientSecret
	}
	next.ClientSecretSet = false
	next.ClearClientSecret = false
	return next
}

func validateSettings(settings models.SystemSettings) error {
	if err := validateClipboardSettings(settings.Clipboard); err != nil {
		return err
	}
	if !hasAvailableLogin(settings.Auth) {
		return errors.New("at least one login method must be available")
	}
	return nil
}

func validateClipboardSettings(settings models.ClipboardSettings) error {
	switch settings.ExpirationUnit {
	case models.ClipboardExpirationUnitMinute, models.ClipboardExpirationUnitHour, models.ClipboardExpirationUnitDay:
		if settings.ExpirationValue <= 0 {
			return errors.New("clipboard expiration value must be greater than zero")
		}
	case models.ClipboardExpirationUnitNever:
		return nil
	default:
		return errors.New("clipboard expiration unit is invalid")
	}
	return nil
}

func hasAvailableLogin(settings models.AuthSettings) bool {
	return settings.PasswordLoginEnabled ||
		providerConfigured(settings.Google) ||
		providerConfigured(settings.GitHub)
}

func providerConfigured(provider models.OAuthProviderConfig) bool {
	return provider.Enabled &&
		strings.TrimSpace(provider.ClientID) != "" &&
		strings.TrimSpace(provider.ClientSecret) != ""
}

func sanitizeProviderConfig(config models.OAuthProviderConfig) models.OAuthProviderConfig {
	config.ClientSecretSet = strings.TrimSpace(config.ClientSecret) != ""
	config.ClientSecret = ""
	config.ClearClientSecret = false
	return config
}

func cloneSettings(settings models.SystemSettings) models.SystemSettings {
	settings.Auth.AllowedEmailDomains = append([]string(nil), settings.Auth.AllowedEmailDomains...)
	return settings
}

func normalizeEmailDomains(domains []string) []string {
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
