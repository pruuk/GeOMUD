package web

import (
	"net/http"
	"sync"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var (
	authCacheMu = sync.RWMutex{}
	// authCache maps the Authorization header value to its expiry time.
	authCache = map[string]time.Time{}
	// authUserCache maps the Authorization header value to the loaded Account.
	authUserCache = map[string]*accounts.Account{}
)

func handlerToHandlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// doBasicAuth authenticates the request using HTTP Basic Auth. Both the
// "admin" and "mod" roles are accepted. The authenticated Account is stored
// in the request context so downstream handlers can access it via
// GetAuthedUser(r).
func doBasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if IsInternalRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")

		authCacheMu.RLock()
		expiry, cached := authCache[authHeader]
		cachedUser := authUserCache[authHeader]
		authCacheMu.RUnlock()

		if cached && expiry.After(time.Now()) && cachedUser != nil {
			r = r.WithContext(withAuthedUser(r.Context(), cachedUser))
			next.ServeHTTP(w, r)
			return
		}

		// Evict stale cache entry if present.
		if cached {
			authCacheMu.Lock()
			delete(authCache, authHeader)
			delete(authUserCache, authHeader)
			authCacheMu.Unlock()
		}

		username, password, ok := r.BasicAuth()
		if ok {
			uRecord, err := accounts.LoadUser(username, true)
			if err == nil && uRecord.PasswordMatches(password) {
				if uRecord.Role == accounts.RoleAdmin || uRecord.Role == accounts.RoleMod {

					mudlog.Warn("ADMIN LOGIN", "username", username, "role", uRecord.Role, "success", true)

					authCacheMu.Lock()
					authCache[authHeader] = time.Now().Add(time.Minute * 30)
					authUserCache[authHeader] = uRecord
					authCacheMu.Unlock()

					r = r.WithContext(withAuthedUser(r.Context(), uRecord))
					next.ServeHTTP(w, r)
					return

				} else {
					mudlog.Error("ADMIN LOGIN", "username", username, "success", false, "error", "Role="+uRecord.Role)
				}
			} else if err != nil {
				mudlog.Error("ADMIN LOGIN", "username", username, "success", false, "error", err)
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
