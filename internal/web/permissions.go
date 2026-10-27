package web

import (
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
)

// RequirePermission wraps a handler so that only users with the given
// permission key (or the admin role) may proceed. For API routes the caller
// should use this after doBasicAuth so that GetAuthedUser is available.
//
// When the request is an internal request it bypasses the check entirely.
//
// API routes (paths starting with /admin/api/) receive a JSON 403 response.
// All other routes receive an admin-themed 403 page showing the required key.
func RequirePermission(permKey string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsInternalRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		u := GetAuthedUser(r)
		if u == nil || !u.HasPermission(permKey) {
			writeForbidden(w, r, permKey)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAnyPermission is like RequirePermission but passes when the user
// holds at least one of the supplied permission keys.
func RequireAnyPermission(permKeys []string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsInternalRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		u := GetAuthedUser(r)
		if u != nil {
			for _, key := range permKeys {
				if u.HasPermission(key) {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		writeForbidden(w, r, strings.Join(permKeys, " or "))
	})
}

// writeForbidden sends an appropriate 403 response. API paths get JSON;
// browser paths get an admin-themed 403 page that shows the required permission.
func writeForbidden(w http.ResponseWriter, r *http.Request, requiredPerm string) {
	if strings.HasPrefix(r.URL.Path, "/admin/api/") {
		writeAPIError(w, http.StatusForbidden, "forbidden: requires permission "+requiredPerm)
		return
	}

	w.WriteHeader(http.StatusForbidden)
	serveAdminTemplate(w, r, "403.html", map[string]any{
		"REQUIRED_PERMISSION": requiredPerm,
	})
}

// RequireAdmin wraps a handler so that only users with the admin role may
// proceed. Unlike RequirePermission, mods are never granted access regardless
// of their permission set.
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsInternalRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		u := GetAuthedUser(r)
		if u == nil || u.Role != accounts.RoleAdmin {
			writeForbidden(w, r, "admin role required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authedUserIsAdmin returns true when the request carries an admin-role user.
// Used by page handlers to decide whether to show admin-only UI elements.
func authedUserIsAdmin(r *http.Request) bool {
	u := GetAuthedUser(r)
	return u != nil && u.Role == accounts.RoleAdmin
}
