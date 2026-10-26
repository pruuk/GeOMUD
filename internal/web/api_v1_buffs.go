package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/buffs"
)

// GET /admin/api/v1/buffs
// Returns all buff flags (flag specs) AND all buff specs.
func apiV1GetBuffs(w http.ResponseWriter, r *http.Request) {
	type buffsResponse struct {
		Flags []buffs.FlagSpec        `json:"flags"`
		Specs map[int]*buffs.BuffSpec `json:"specs"`
	}
	writeJSON(w, http.StatusOK, APIResponse[buffsResponse]{
		Success: true,
		Data: buffsResponse{
			Flags: buffs.GetAllFlagSpecsSorted(),
			Specs: buffs.GetAllBuffSpecs(),
		},
	})
}

// resolveBuffId converts a path segment to an integer buff ID and writes an
// error response if it is not a valid integer or the buff does not exist.
func resolveBuffId(w http.ResponseWriter, idStr string) (int, bool) {
	buffId, err := strconv.Atoi(idStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "buffId must be an integer: "+idStr)
		return 0, false
	}
	if buffs.GetBuffSpec(buffId) == nil {
		writeAPIError(w, http.StatusNotFound, "buff not found: "+idStr)
		return 0, false
	}
	return buffId, true
}

// GET /admin/api/v1/buffs/{buffId}
func apiV1GetBuff(w http.ResponseWriter, r *http.Request) {
	buffId, ok := resolveBuffId(w, r.PathValue("buffId"))
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*buffs.BuffSpec]{
		Success: true,
		Data:    buffs.GetBuffSpec(buffId),
	})
}

// PATCH /admin/api/v1/buffs/{buffId}
func apiV1PatchBuff(w http.ResponseWriter, r *http.Request) {
	buffId, ok := resolveBuffId(w, r.PathValue("buffId"))
	if !ok {
		return
	}

	existing := buffs.GetBuffSpec(buffId)

	// Decode into a copy so only supplied fields are changed.
	updated := *existing
	// Nil out map fields that must be fully replaced (not merged) by the JSON
	// decode. Go's json.Decoder merges into existing maps, so without this a
	// stat mod removed in the editor (zeroed/deleted, hence omitted from the
	// payload) would keep its old value and reappear on reload.
	updated.StatMods = nil
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}

	// Preserve the canonical ID.
	updated.BuffId = buffId

	if err := buffs.SaveBuffSpec(&updated); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[*buffs.BuffSpec]{
		Success: true,
		Data:    &updated,
	})
}

// DELETE /admin/api/v1/buffs/{buffId}
func apiV1DeleteBuff(w http.ResponseWriter, r *http.Request) {
	buffId, ok := resolveBuffId(w, r.PathValue("buffId"))
	if !ok {
		return
	}

	if err := buffs.DeleteBuffSpec(buffId); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
