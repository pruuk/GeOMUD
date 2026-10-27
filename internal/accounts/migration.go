package accounts

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Note: alt character file migration (renaming <username>-alts.yaml to
// <userid>.alts.yaml) is handled by the alt-characters module, not here.

var (
	ErrBothFilesExist = errors.New("could not migrate due to both file formats existing")
)

//
// Handles user file migration
//

func DoUserMigrations() {

	DoFilenameMigrationV1()

}

// giveBodyId makes u's character its body ID if it has none, and reports
// whether it did. A body ID is made once and never replaced (Phase 4b
// decision 11), so this is only for a new character and for characters saved
// before body IDs existed.
func giveBodyId(u *Account) bool {
	if u.Character == nil || !u.Character.Id.IsZero() {
		return false
	}
	u.Character.Id = ids.New(ids.NoCell)
	return true
}

// DoBodyIdMigrationV1 gives every saved character without a body ID one and
// saves it. main runs it at startup before anyone can log in, so no two loads
// of one account can race to give it two IDs. It is idempotent.
func DoBodyIdMigrationV1() error {
	var errorResult error
	SearchOfflineUsers(func(u *Account) bool {
		if giveBodyId(u) {
			if err := SaveUser(*u); err != nil {
				errorResult = err
				return false
			}
		}
		return true
	})
	return errorResult
}

func DoFilenameMigrationV1() error {

	var errorResult error = nil

	SearchOfflineUsers(func(u *Account) bool {

		oldUserPath := util.FilePath(string(configs.GetFilePathsConfig().DataFiles), `/`, `users`, `/`, strings.ToLower(u.Username)+`.yaml`)
		newUserPath := util.FilePath(string(configs.GetFilePathsConfig().DataFiles), `/`, `users`, `/`, strconv.Itoa(u.UserId)+`.yaml`)

		_, err := os.Stat(oldUserPath)
		oldUserPathExists := err == nil

		_, err = os.Stat(newUserPath)
		newUserPathExists := err == nil

		if oldUserPathExists && newUserPathExists {
			errorResult = ErrBothFilesExist
			return false
		}

		if oldUserPathExists {
			err := os.Rename(oldUserPath, newUserPath)
			if err != nil {
				errorResult = err
				return false
			}
		}

		return true
	})

	return errorResult
}
