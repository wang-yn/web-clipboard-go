package services

import "testing"

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
