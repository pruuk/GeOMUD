package web

import (
	"context"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/accounts"
)

type contextKey int

const (
	internalCallerKey contextKey = iota
	testModeKey
	authedUserKey
)

// withInternalContext returns a copy of ctx marked as an internal request.
func withInternalContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internalCallerKey, true)
}

// IsInternalRequest reports whether r was dispatched via InternalRequest rather
// than arriving over the network. Handlers can use this to skip audit logging,
// rate limiting, or other concerns that only apply to external callers.
func IsInternalRequest(r *http.Request) bool {
	v, _ := r.Context().Value(internalCallerKey).(bool)
	return v
}

// withTestModeContext returns a copy of ctx marked as a test-mode request.
func withTestModeContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, testModeKey, true)
}

// IsTestModeRequest reports whether r was sent with X-Test-Mode: true.
func IsTestModeRequest(r *http.Request) bool {
	v, _ := r.Context().Value(testModeKey).(bool)
	return v
}

// withAuthedUser stores the authenticated UserRecord in the request context.
func withAuthedUser(ctx context.Context, u *accounts.Account) context.Context {
	return context.WithValue(ctx, authedUserKey, u)
}

// GetAuthedUser retrieves the authenticated UserRecord from the request context.
// Returns nil when no user is stored (e.g. internal requests).
func GetAuthedUser(r *http.Request) *accounts.Account {
	u, _ := r.Context().Value(authedUserKey).(*accounts.Account)
	return u
}
