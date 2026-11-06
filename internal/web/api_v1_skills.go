package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/skills"
)

// resolveSkillId lowercases a path segment and writes a 404 response if the
// skill is not loaded.
func resolveSkillId(w http.ResponseWriter, idStr string) (string, bool) {
	skillId := strings.ToLower(strings.TrimSpace(idStr))
	if !skills.SkillExists(skillId) {
		writeAPIError(w, http.StatusNotFound, "skill not found: "+idStr)
		return "", false
	}
	return skillId, true
}

// GET /admin/api/v1/skills
func apiV1GetSkills(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse[map[string]*skills.Skill]{
		Success: true,
		Data:    skills.GetSkillsMap(),
	})
}

// POST /admin/api/v1/skills
func apiV1CreateSkill(w http.ResponseWriter, r *http.Request) {
	var newSkill skills.Skill
	if err := json.NewDecoder(r.Body).Decode(&newSkill); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	if err := skills.CreateSkill(&newSkill); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[*skills.Skill]{Success: true, Data: &newSkill})
}

// PATCH /admin/api/v1/skills/{skillId}
func apiV1PatchSkill(w http.ResponseWriter, r *http.Request) {
	skillId, ok := resolveSkillId(w, r.PathValue("skillId"))
	if !ok {
		return
	}
	existing := skills.GetSkill(skillId)
	updated := *existing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
	updated.SkillId = skillId // id is immutable
	if err := skills.SaveSkill(&updated); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[*skills.Skill]{Success: true, Data: &updated})
}

// DELETE /admin/api/v1/skills/{skillId}
func apiV1DeleteSkill(w http.ResponseWriter, r *http.Request) {
	skillId, ok := resolveSkillId(w, r.PathValue("skillId"))
	if !ok {
		return
	}
	if err := skills.DeleteSkill(skillId); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}
