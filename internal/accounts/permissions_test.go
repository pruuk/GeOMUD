package accounts

import (
	"testing"
)

func newModUser(permissions ...string) *Account {
	u := &Account{
		Role:        RoleMod,
		Permissions: permissions,
	}
	return u
}

func TestHasRolePermission_Admin(t *testing.T) {
	u := &Account{Role: RoleAdmin}
	if !u.HasRolePermission("anything") {
		t.Error("admin should always have permission")
	}
	if !u.HasRolePermission("room.edit.exits") {
		t.Error("admin should have any permission")
	}
}

func TestHasRolePermission_User(t *testing.T) {
	u := &Account{Role: RoleUser}
	if u.HasRolePermission("mobs.read") {
		t.Error("regular user should have no permissions")
	}
}

func TestHasRolePermission_Guest(t *testing.T) {
	u := &Account{Role: RoleGuest}
	if u.HasRolePermission("mobs.read") {
		t.Error("guest should have no permissions")
	}
}

func TestHasRolePermission_ModExactMatch(t *testing.T) {
	u := newModUser("mobs.read", "rooms.read")
	if !u.HasRolePermission("mobs.read") {
		t.Error("mod should have mobs.read")
	}
	if !u.HasRolePermission("rooms.read") {
		t.Error("mod should have rooms.read")
	}
	if u.HasRolePermission("mobs.write") {
		t.Error("mod should not have mobs.write")
	}
}

func TestHasRolePermission_ModPrefixMatch(t *testing.T) {
	// Granting "room" should satisfy "room.edit", "room.edit.exits", etc.
	u := newModUser("room")
	if !u.HasRolePermission("room.edit") {
		t.Error("room should grant room.edit")
	}
	if !u.HasRolePermission("room.edit.exits") {
		t.Error("room should grant room.edit.exits")
	}
	if !u.HasRolePermission("room") {
		t.Error("room should grant room (exact)")
	}
	// Should NOT grant an unrelated key that starts with "room" but different segment.
	// "roommate" does not start with "room." so it should not match.
	if u.HasRolePermission("roommate") {
		t.Error("room should not grant roommate (no dot boundary)")
	}
}

func TestHasRolePermission_ModSimpleMatch(t *testing.T) {
	// simpleMatch=true: requested "room" should match granted "room.edit"
	u := newModUser("room.edit")
	if !u.HasRolePermission("room", true) {
		t.Error("simpleMatch: room should match granted room.edit")
	}
	// Without simpleMatch, "room" should NOT match granted "room.edit"
	if u.HasRolePermission("room") {
		t.Error("no simpleMatch: room should not match granted room.edit")
	}
}

func TestHasRolePermission_ModNoPrefixFalsePositive(t *testing.T) {
	// "mobs" should not grant "mobsomething" (no dot boundary)
	u := newModUser("mobs")
	if u.HasRolePermission("mobsomething") {
		t.Error("mobs should not grant mobsomething")
	}
}

func TestHasRolePermission_ModEmptyPermissions(t *testing.T) {
	u := newModUser()
	if u.HasRolePermission("mobs.read") {
		t.Error("mod with no permissions should be denied")
	}
}

func TestHasPermission_AdminAlwaysTrue(t *testing.T) {
	u := &Account{Role: RoleAdmin}
	if !u.HasPermission("config.write") {
		t.Error("HasPermission: admin should always return true")
	}
}

func TestHasPermission_ModWithPermission(t *testing.T) {
	u := newModUser("config.read")
	if !u.HasPermission("config.read") {
		t.Error("HasPermission: mod should have config.read")
	}
	if u.HasPermission("config.write") {
		t.Error("HasPermission: mod should not have config.write")
	}
}
