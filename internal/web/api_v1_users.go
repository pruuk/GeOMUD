package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/items"
)

// resolveUserId resolves a path segment that is a numeric user ID.
// Returns 0 and writes an error response if the value is not a valid integer.
func resolveUserId(w http.ResponseWriter, idStr string) int {
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid user id: "+idStr)
		return 0
	}
	return id
}

// loadUserRecord loads an Account by id, checking the online cache first.
// Writes a 404 error response and returns nil when the user does not exist.
func loadUserRecord(w http.ResponseWriter, userId int) *accounts.Account {
	if u := accounts.GetByUserId(userId); u != nil {
		return u
	}
	u, err := accounts.LoadUserById(userId)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "user not found")
		return nil
	}
	return u
}

// GET /admin/api/v1/users/{userid}
func apiV1GetUser(w http.ResponseWriter, r *http.Request) {
	userId := resolveUserId(w, r.PathValue("userid"))
	if userId == 0 {
		return
	}

	u := loadUserRecord(w, userId)
	if u == nil {
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*accounts.Account]{
		Success: true,
		Data:    u,
	})
}

// decodeUserPatch applies an admin PATCH body to a copy of u. The form sends
// the complete desired skills, spellbook, items and equipment, and
// json.Unmarshal merges into existing maps and slice elements, so those start
// empty on the copy: a removed entry stays removed, an item never takes
// another item's fields, and the live record is not touched. Items the form
// kept keep their identity (characters.KeepItemIdentity). A body that leaves
// out the items or the equipment leaves them as they were.
func decodeUserPatch(u *accounts.Account, body []byte) (accounts.Account, error) {
	updated := *u
	if u.Character != nil {
		charCopy := *u.Character
		charCopy.Skills = nil
		charCopy.SpellBook = nil
		charCopy.Items = nil
		charCopy.Equipment = characters.Worn{}
		updated.Character = &charCopy
	}
	if err := json.Unmarshal(body, &updated); err != nil {
		return accounts.Account{}, err
	}
	if u.Character == nil || updated.Character == nil {
		return updated, nil
	}

	var present struct {
		Character *struct {
			Items     json.RawMessage
			Equipment json.RawMessage
		}
	}
	_ = json.Unmarshal(body, &present) // the body already decoded above
	if present.Character == nil || present.Character.Items == nil {
		updated.Character.Items = append([]items.Item(nil), u.Character.Items...)
	}
	if present.Character == nil || present.Character.Equipment == nil {
		updated.Character.Equipment = u.Character.Equipment
	}
	updated.Character.KeepItemIdentity(u.Character)
	return updated, nil
}

// PATCH /admin/api/v1/users/{userid}
func apiV1PatchUser(w http.ResponseWriter, r *http.Request) {
	userId := resolveUserId(w, r.PathValue("userid"))
	if userId == 0 {
		return
	}

	u := loadUserRecord(w, userId)
	if u == nil {
		return
	}

	// Capture the plaintext password from the request before decoding into the
	// Account, because Account.Password stores a bcrypt hash and we need
	// to call SetPassword to hash a new plaintext value.
	var raw struct {
		Password string `json:"Password"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "failed to read request body: "+err.Error())
		return
	}
	_ = json.Unmarshal(body, &raw)

	updated, err := decodeUserPatch(u, body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	// Preserve the canonical ID; callers cannot change it via PATCH.
	updated.UserId = userId

	// If a plaintext password was supplied, hash it properly.
	if raw.Password != "" && !isBcryptHash(raw.Password) {
		if err := updated.SetPassword(raw.Password); err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid password: "+err.Error())
			return
		}
	}

	if updated.Character.Gold < 0 {
		updated.Character.Gold = 0
	}
	if updated.Character.Bank < 0 {
		updated.Character.Bank = 0
	}

	updated.Character.Validate()

	if err := accounts.SaveUser(updated); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	accounts.UpdateOnlineUser(updated)

	writeJSON(w, http.StatusOK, APIResponse[*accounts.Account]{
		Success: true,
		Data:    &updated,
	})
}

// isBcryptHash returns true when s already looks like a bcrypt hash so we do
// not double-hash a value that was round-tripped through the API.
func isBcryptHash(s string) bool {
	return len(s) > 4 && (s[:4] == "$2a$" || s[:4] == "$2b$")
}

// POST /admin/api/v1/users
func apiV1CreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"Username"`
		Password string `json:"Password"`
		Role     string `json:"Role"`
		Email    string `json:"EmailAddress"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	if body.Username == "" {
		writeAPIError(w, http.StatusBadRequest, "Username is required")
		return
	}

	if accounts.Exists(body.Username) {
		writeAPIError(w, http.StatusConflict, "username already exists")
		return
	}

	u := accounts.NewAccount(0, 0)
	if err := u.SetUsername(body.Username); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.Password != "" {
		if err := u.SetPassword(body.Password); err != nil {
			writeAPIError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.Role != "" {
		u.Role = body.Role
	}
	if body.Email != "" {
		u.EmailAddress = body.Email
	}

	if err := accounts.CreateUser(u); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse[*accounts.Account]{
		Success: true,
		Data:    u,
	})
}

// GET /admin/api/v1/users/search
//
// Query params (mutually exclusive; name takes priority):
//
//	name=<string>  - search by username or numeric user ID (existing behaviour)
//	role=<string>  - return all users with the given role (admin, mod, user, guest)
func apiV1SearchUsers(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	role := r.URL.Query().Get("role")

	if name == "" && role == "" {
		writeAPIError(w, http.StatusBadRequest, "name or role query parameter is required")
		return
	}

	var results []accounts.UserSearchResult
	if name != "" {
		results = accounts.SearchUsers(name)
	} else {
		results = accounts.SearchUsersByRole(role)
	}
	if results == nil {
		results = []accounts.UserSearchResult{}
	}

	writeJSON(w, http.StatusOK, APIResponse[[]accounts.UserSearchResult]{
		Success: true,
		Data:    results,
	})
}
