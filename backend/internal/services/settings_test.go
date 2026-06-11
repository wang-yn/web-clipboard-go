package services

import (
	"path/filepath"
	"testing"
	"time"

	"web-clipboard-go/backend/internal/models"
)

func TestSettingsServiceDefaultsAndPersistence(t *testing.T) {
	service, err := NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	settings := service.GetSettings()
	if !settings.Auth.PasswordLoginEnabled {
		t.Fatal("password login must be enabled by default")
	}
	if settings.Clipboard.ExpirationValue != 10 || settings.Clipboard.ExpirationUnit != models.ClipboardExpirationUnitMinute {
		t.Fatalf("unexpected default clipboard expiration: %#v", settings.Clipboard)
	}

	settings.Auth.Google.Enabled = true
	settings.Auth.Google.ClientID = "google-client"
	settings.Auth.Google.ClientSecret = "google-secret"
	settings.Clipboard.ExpirationValue = 2
	settings.Clipboard.ExpirationUnit = models.ClipboardExpirationUnitHour
	if err := service.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}

	reloaded, err := NewSettingsService(filepath.Dir(service.filePath))
	if err != nil {
		t.Fatal(err)
	}
	loaded := reloaded.GetSettings()
	if !loaded.Auth.Google.Enabled || loaded.Auth.Google.ClientSecret != "google-secret" {
		t.Fatalf("saved oauth settings were not reloaded: %#v", loaded.Auth.Google)
	}
	if loaded.Clipboard.ExpirationValue != 2 || loaded.Clipboard.ExpirationUnit != models.ClipboardExpirationUnitHour {
		t.Fatalf("saved clipboard settings were not reloaded: %#v", loaded.Clipboard)
	}
}

func TestSettingsServiceSanitizedResponseDoesNotExposeSecrets(t *testing.T) {
	service, err := NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	settings := service.GetSettings()
	settings.Auth.Google.Enabled = true
	settings.Auth.Google.ClientID = "google-client"
	settings.Auth.Google.ClientSecret = "google-secret"
	if err := service.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}

	response := service.GetSettingsResponse()
	if response.Auth.Google.ClientSecret != "" {
		t.Fatal("settings response must not expose google client secret")
	}
	if !response.Auth.Google.ClientSecretSet {
		t.Fatal("settings response must indicate that google client secret is configured")
	}
}

func TestSettingsServiceRejectsNoAvailableLoginMethod(t *testing.T) {
	service, err := NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	settings := service.GetSettings()
	settings.Auth.PasswordLoginEnabled = false
	settings.Auth.Google.Enabled = true
	settings.Auth.Google.ClientID = "google-client"
	settings.Auth.Google.ClientSecret = ""
	settings.Auth.GitHub.Enabled = false

	if err := service.SaveSettings(settings); err == nil {
		t.Fatal("expected settings without available login method to be rejected")
	}
}

func TestSettingsServiceKeepsCurrentSettingsWhenWriteFails(t *testing.T) {
	service, err := NewSettingsService(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	original := service.GetSettings()

	service.filePath = filepath.Join(t.TempDir(), "missing", "settings.json")
	next := original
	next.Clipboard.ExpirationValue = 5
	next.Clipboard.ExpirationUnit = models.ClipboardExpirationUnitDay

	if err := service.SaveSettings(next); err == nil {
		t.Fatal("expected save to fail when settings file cannot be written")
	}

	current := service.GetSettings()
	if current.Clipboard.ExpirationValue != original.Clipboard.ExpirationValue ||
		current.Clipboard.ExpirationUnit != original.Clipboard.ExpirationUnit {
		t.Fatalf("settings changed in memory after failed write: %#v", current.Clipboard)
	}
}

func TestClipboardExpirationFromSettings(t *testing.T) {
	now := time.Date(2026, 6, 11, 8, 0, 0, 0, time.UTC)
	cases := []struct {
		name     string
		settings models.ClipboardSettings
		want     time.Time
	}{
		{
			name:     "minutes",
			settings: models.ClipboardSettings{ExpirationValue: 15, ExpirationUnit: models.ClipboardExpirationUnitMinute},
			want:     now.Add(15 * time.Minute),
		},
		{
			name:     "hours",
			settings: models.ClipboardSettings{ExpirationValue: 2, ExpirationUnit: models.ClipboardExpirationUnitHour},
			want:     now.Add(2 * time.Hour),
		},
		{
			name:     "days",
			settings: models.ClipboardSettings{ExpirationValue: 3, ExpirationUnit: models.ClipboardExpirationUnitDay},
			want:     now.Add(72 * time.Hour),
		},
		{
			name:     "never",
			settings: models.ClipboardSettings{ExpirationValue: 1, ExpirationUnit: models.ClipboardExpirationUnitNever},
			want:     time.Time{},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.settings.ExpiresAt(now); !got.Equal(tt.want) {
				t.Fatalf("ExpiresAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClipboardItemExpiredTreatsZeroExpiryAsNever(t *testing.T) {
	now := time.Date(2026, 6, 11, 8, 0, 0, 0, time.UTC)
	if models.ClipboardItemExpired(&models.ClipboardItem{ExpiresAt: time.Time{}}, now) {
		t.Fatal("zero expiresAt must mean never expires")
	}
	if !models.ClipboardItemExpired(&models.ClipboardItem{ExpiresAt: now.Add(-time.Second)}, now) {
		t.Fatal("past expiresAt must be expired")
	}
	if models.ClipboardItemExpired(&models.ClipboardItem{ExpiresAt: now.Add(time.Second)}, now) {
		t.Fatal("future expiresAt must not be expired")
	}
}
