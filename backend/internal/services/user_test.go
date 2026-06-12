package services

import (
	"regexp"
	"strings"
	"testing"
)

func TestDefaultAdminUsesRandomInitialPassword(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	admin := manager.GetUserByUsername("admin")
	if admin == nil {
		t.Fatal("default admin missing")
	}
	if _, err := manager.ValidateCredentials("admin", "admin123"); err == nil {
		t.Fatal("default admin must not use the old fixed admin123 password")
	}
}

func TestGenerateInitialAdminPasswordUsesConsoleSafeRandomFormat(t *testing.T) {
	password, err := generateInitialAdminPassword()
	if err != nil {
		t.Fatal(err)
	}
	if len(password) < 20 {
		t.Fatalf("initial admin password is too short: %d", len(password))
	}
	if strings.ContainsAny(password, "\"'`\\") {
		t.Fatalf("initial admin password contains shell-confusing characters: %q", password)
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(password) {
		t.Fatalf("initial admin password should be URL-safe, got %q", password)
	}
}

func TestUpdateUserRejectsDisablingLastActiveAdmin(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	admin := manager.GetUserByUsername("admin")
	if admin == nil {
		t.Fatal("default admin missing")
	}

	inactive := false
	if _, err := manager.UpdateUser(admin.ID, "", "", &inactive); err == nil {
		t.Fatal("expected disabling the last active admin to fail")
	}

	if updated := manager.GetUser(admin.ID); !updated.IsActive || updated.Role != "admin" {
		t.Fatalf("last admin was changed after rejected update: role=%s active=%v", updated.Role, updated.IsActive)
	}
}

func TestUpdateUserRejectsDemotingLastActiveAdmin(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	admin := manager.GetUserByUsername("admin")
	if admin == nil {
		t.Fatal("default admin missing")
	}

	if _, err := manager.UpdateUser(admin.ID, "", "user", nil); err == nil {
		t.Fatal("expected demoting the last active admin to fail")
	}

	if updated := manager.GetUser(admin.ID); updated.Role != "admin" || !updated.IsActive {
		t.Fatalf("last admin was changed after rejected update: role=%s active=%v", updated.Role, updated.IsActive)
	}
}

func TestCreateExternalUserStoresIdentityAndRejectsPasswordLogin(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	identity := ExternalIdentity{
		Provider:      "github",
		Subject:       "12345",
		Email:         "octo@example.com",
		EmailVerified: true,
		Username:      "octocat",
		DisplayName:   "Octo Cat",
		AvatarURL:     "https://example.com/avatar.png",
	}

	user, err := manager.CreateExternalUser(identity, "user")
	if err != nil {
		t.Fatal(err)
	}
	if user.Password != "" {
		t.Fatal("external users must not receive a local password hash")
	}
	if user.Role != "user" {
		t.Fatalf("external users must default to user role, got %q", user.Role)
	}
	if len(user.Identities) != 1 {
		t.Fatalf("expected one linked identity, got %d", len(user.Identities))
	}

	linked := manager.GetUserByExternalIdentity("github", "12345")
	if linked == nil || linked.ID != user.ID {
		t.Fatalf("external identity lookup returned %#v, want user %s", linked, user.ID)
	}

	if _, err := manager.ValidateCredentials(user.Username, "anything"); err == nil {
		t.Fatal("external user without password must not be able to use local password login")
	}
}

func TestLinkExternalIdentityRejectsDuplicateProviderSubject(t *testing.T) {
	manager, err := NewUserManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	first, err := manager.CreateExternalUser(ExternalIdentity{
		Provider:      "google",
		Subject:       "subject-1",
		Email:         "first@example.com",
		EmailVerified: true,
		Username:      "first",
	}, "user")
	if err != nil {
		t.Fatal(err)
	}

	second, err := manager.CreateUser("second", "secret123", "second@example.com", "user")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := manager.LinkExternalIdentity(second.ID, ExternalIdentity{
		Provider:      "google",
		Subject:       "subject-1",
		Email:         "second@example.com",
		EmailVerified: true,
		Username:      "second",
	}); err == nil {
		t.Fatal("expected duplicate external identity to be rejected")
	}

	linked := manager.GetUserByExternalIdentity("google", "subject-1")
	if linked == nil || linked.ID != first.ID {
		t.Fatalf("duplicate link changed existing identity owner: %#v", linked)
	}
}
