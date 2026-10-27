package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
)

func TestRequirePermission_AdminAlwaysPasses(t *testing.T) {
	admin := &accounts.Account{Role: accounts.RoleAdmin}
	req := httptest.NewRequest(http.MethodGet, "/admin/mobs", nil)
	req = req.WithContext(withAuthedUser(req.Context(), admin))

	called := false
	handler := RequirePermission("mobs.read", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("admin: got %d, want 200", rec.Code)
	}
	if !called {
		t.Fatal("admin: handler should have been called")
	}
}

func TestRequirePermission_ModWithPermissionPasses(t *testing.T) {
	mod := &accounts.Account{Role: accounts.RoleMod, Permissions: []string{"mobs.read"}}
	req := httptest.NewRequest(http.MethodGet, "/admin/mobs", nil)
	req = req.WithContext(withAuthedUser(req.Context(), mod))

	called := false
	handler := RequirePermission("mobs.read", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("mod with perm: got %d, want 200", rec.Code)
	}
	if !called {
		t.Fatal("mod with perm: handler should have been called")
	}
}

func TestRequirePermission_ModWithoutPermissionForbidden(t *testing.T) {
	mod := &accounts.Account{Role: accounts.RoleMod, Permissions: []string{"rooms.read"}}
	// Use an API path so writeForbidden returns JSON without needing template files.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/mobs", nil)
	req = req.WithContext(withAuthedUser(req.Context(), mod))

	called := false
	handler := RequirePermission("mobs.read", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("mod without perm: got %d, want 403", rec.Code)
	}
	if called {
		t.Fatal("mod without perm: handler should not have been called")
	}
}

func TestRequirePermission_APIRouteReturnJSON(t *testing.T) {
	mod := &accounts.Account{Role: accounts.RoleMod, Permissions: []string{}}
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/mobs", nil)
	req = req.WithContext(withAuthedUser(req.Context(), mod))

	handler := RequirePermission("mobs.read", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("api route: got %d, want 403", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("api route: Content-Type = %q, want application/json", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "mobs.read") {
		t.Fatalf("api route: body should contain permission key, got: %s", body)
	}
}

func TestRequirePermission_NoAuthedUserForbidden(t *testing.T) {
	// Use an API path so writeForbidden returns JSON without needing template files.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/mobs", nil)
	// No authed user in context.

	called := false
	handler := RequirePermission("mobs.read", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("no authed user: got %d, want 403", rec.Code)
	}
	if called {
		t.Fatal("no authed user: handler should not have been called")
	}
}

func TestRequirePermission_InternalRequestBypasses(t *testing.T) {
	// Internal requests skip permission checks.
	req := httptest.NewRequest(http.MethodGet, "/admin/mobs", nil)
	req = req.WithContext(withInternalContext(req.Context()))

	called := false
	handler := RequirePermission("mobs.read", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("internal request: got %d, want 200", rec.Code)
	}
	if !called {
		t.Fatal("internal request: handler should have been called")
	}
}

func TestRequirePermission_ModPrefixGrantPasses(t *testing.T) {
	// Granting "room" should satisfy "room.edit.exits" via prefix match.
	mod := &accounts.Account{Role: accounts.RoleMod, Permissions: []string{"room"}}
	// Use an API path so the handler path is clean.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/rooms", nil)
	req = req.WithContext(withAuthedUser(req.Context(), mod))

	called := false
	handler := RequirePermission("room.edit.exits", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("prefix grant: got %d, want 200", rec.Code)
	}
	if !called {
		t.Fatal("prefix grant: handler should have been called")
	}
}
