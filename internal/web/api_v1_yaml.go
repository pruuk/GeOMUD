package web

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mutators"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/spells"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// yamlResponse is the standard payload for all YAML viewer endpoints.
type yamlResponse struct {
	YAML string `json:"yaml"`
	Path string `json:"path"`
}

// readYAMLFile reads a file from disk and returns its content as a string.
// Writes an appropriate error response and returns ("", false) on failure.
func readYAMLFile(w http.ResponseWriter, path string) (string, bool) {
	path = util.FilePath(path)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeAPIError(w, http.StatusNotFound, "yaml file not found: "+path)
		} else {
			writeAPIError(w, http.StatusInternalServerError, "reading yaml file: "+err.Error())
		}
		return "", false
	}
	return string(data), true
}

func dataFiles() string {
	return configs.GetFilePathsConfig().DataFiles.String()
}

// relPath strips the DataFiles prefix from an absolute path and returns the
// portion relative to the world directory (e.g. "buffs/0-meditating.yaml").
// Forward slashes are used regardless of OS.
func relPath(absPath string) string {
	base := dataFiles()
	rel := strings.TrimPrefix(
		strings.ReplaceAll(absPath, `\`, `/`),
		strings.ReplaceAll(base, `\`, `/`),
	)
	return strings.TrimPrefix(rel, "/")
}

// yamlResp builds a yamlResponse with the content and its relative path.
func yamlResp(content, absPath string) yamlResponse {
	return yamlResponse{YAML: content, Path: relPath(absPath)}
}

// ---------------------------------------------------------------------------
// Buffs
// ---------------------------------------------------------------------------

// GET /admin/api/v1/buffs/{buffId}/yaml
func apiV1GetBuffYAML(w http.ResponseWriter, r *http.Request) {
	buffId, ok := resolveBuffId(w, r.PathValue("buffId"))
	if !ok {
		return
	}
	spec := buffs.GetBuffSpec(buffId)
	path := dataFiles() + `/buffs/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Mobs
// ---------------------------------------------------------------------------

// GET /admin/api/v1/mobs/{mobId}/yaml
func apiV1GetMobYAML(w http.ResponseWriter, r *http.Request) {
	mobId := resolveMobId(w, r.PathValue("mobId"))
	if mobId == 0 {
		return
	}
	spec := mobs.GetMobSpec(mobId)
	path := dataFiles() + `/mobs/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Items
// ---------------------------------------------------------------------------

// GET /admin/api/v1/items/{itemId}/yaml
func apiV1GetItemYAML(w http.ResponseWriter, r *http.Request) {
	itemId := resolveItemId(w, r.PathValue("itemId"))
	if itemId == 0 {
		return
	}
	spec := items.GetItemSpec(itemId)
	path := dataFiles() + `/items/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Spells
// ---------------------------------------------------------------------------

// GET /admin/api/v2/spells/{spellId}/yaml
func apiV2GetSpellYAML(w http.ResponseWriter, r *http.Request) {
	spellId := r.PathValue("spellId")
	spec := spells.GetSpell(spellId)
	if spec == nil {
		writeAPIError(w, http.StatusNotFound, "spell not found: "+spellId)
		return
	}
	path := dataFiles() + `/spells/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Races
// ---------------------------------------------------------------------------

// GET /admin/api/v1/races/{raceId}/yaml
func apiV1GetRaceYAML(w http.ResponseWriter, r *http.Request) {
	raceId, ok := resolveRaceId(w, r.PathValue("raceId"))
	if !ok {
		return
	}
	race := races.GetRace(raceId)
	path := dataFiles() + `/races/` + race.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Skills & Professions
// ---------------------------------------------------------------------------

// GET /admin/api/v1/skills/{skillId}/yaml
func apiV1GetSkillYAML(w http.ResponseWriter, r *http.Request) {
	skillId, ok := resolveSkillId(w, r.PathValue("skillId"))
	if !ok {
		return
	}
	spec := skills.GetSkill(skillId)
	path := dataFiles() + `/skills/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// GET /admin/api/v1/professions/{professionId}/yaml
func apiV1GetProfessionYAML(w http.ResponseWriter, r *http.Request) {
	professionId, ok := resolveProfessionId(w, r.PathValue("professionId"))
	if !ok {
		return
	}
	spec := skills.GetProfessionSpec(professionId)
	path := dataFiles() + `/professions/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Mutators
// ---------------------------------------------------------------------------

// GET /admin/api/v1/mutators/{mutatorId}/yaml
func apiV1GetMutatorYAML(w http.ResponseWriter, r *http.Request) {
	mutatorId := r.PathValue("mutatorId")
	spec := mutators.GetMutatorSpec(mutatorId)
	if spec == nil {
		writeAPIError(w, http.StatusNotFound, "mutator not found: "+mutatorId)
		return
	}
	path := dataFiles() + `/mutators/` + spec.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Biomes
// ---------------------------------------------------------------------------

// GET /admin/api/v1/biomes/{biomeId}/yaml
func apiV1GetBiomeYAML(w http.ResponseWriter, r *http.Request) {
	biomeId := strings.TrimSpace(r.PathValue("biomeId"))
	if biomeId == "" {
		writeAPIError(w, http.StatusBadRequest, "biomeId is required")
		return
	}
	biome, ok := rooms.GetBiome(biomeId)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "biome not found: "+biomeId)
		return
	}
	path := dataFiles() + `/biomes/` + biome.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Rooms
// ---------------------------------------------------------------------------

// GET /admin/api/v1/rooms/{roomId}/yaml
func apiV1GetRoomYAML(w http.ResponseWriter, r *http.Request) {
	roomIdStr := r.PathValue("roomId")
	roomId, err := strconv.Atoi(roomIdStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "roomId must be an integer")
		return
	}
	filePath := rooms.GetRoomTemplatePath(roomId)
	if filePath == "" {
		writeAPIError(w, http.StatusNotFound, fmt.Sprintf("room file not found for room %d", roomId))
		return
	}
	content, ok := readYAMLFile(w, filePath)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, filePath),
	})
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

// GET /admin/api/v1/users/{userid}/yaml
func apiV1GetUserYAML(w http.ResponseWriter, r *http.Request) {
	userId := resolveUserId(w, r.PathValue("userid"))
	if userId == 0 {
		return
	}
	if u := loadUserRecord(w, userId); u == nil {
		return
	}
	path := util.FilePath(dataFiles(), `/`, `users`, `/`, strconv.Itoa(userId)+`.yaml`)
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Keywords (single file)
// ---------------------------------------------------------------------------

// GET /admin/api/v1/keywords/yaml
func apiV1GetKeywordsYAML(w http.ResponseWriter, r *http.Request) {
	path := dataFiles() + `/keywords.yaml`
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Color patterns (single file)
// ---------------------------------------------------------------------------

// GET /admin/api/v1/colorpatterns/yaml
func apiV1GetColorPatternsYAML(w http.ResponseWriter, r *http.Request) {
	path := dataFiles() + `/color-patterns.yaml`
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Color aliases (single file)
// ---------------------------------------------------------------------------

// GET /admin/api/v1/color-aliases/yaml
func apiV1GetColorAliasesYAML(w http.ResponseWriter, r *http.Request) {
	path := dataFiles() + `/ansi-aliases.yaml`
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Attack messages (per-subtype file)
// ---------------------------------------------------------------------------

// GET /admin/api/v1/items/attack-messages/{subtype}/yaml
func apiV1GetAttackMessagesYAML(w http.ResponseWriter, r *http.Request) {
	subtype := strings.TrimSpace(r.PathValue("subtype"))
	if subtype == "" {
		writeAPIError(w, http.StatusBadRequest, "subtype is required")
		return
	}
	msgs := items.GetAllAttackMessages()
	group, found := msgs[items.ItemSubType(subtype)]
	if !found || group == nil {
		writeAPIError(w, http.StatusNotFound, "attack message subtype not found: "+subtype)
		return
	}
	path := dataFiles() + `/combat-messages/` + group.Filepath()
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}

// ---------------------------------------------------------------------------
// Gametime (single file, all calendars)
// ---------------------------------------------------------------------------

// GET /admin/api/v1/gametime/yaml
func apiV1GetGameTimeYAML(w http.ResponseWriter, r *http.Request) {
	path := dataFiles() + `/gametime.yaml`
	content, ok := readYAMLFile(w, path)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, APIResponse[yamlResponse]{
		Success: true,
		Data:    yamlResp(content, path),
	})
}
