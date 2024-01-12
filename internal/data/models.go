package data

import (
	"database/sql"
	"errors"
)

var (
	RecordNotFoundError = errors.New("record not found")
	EditConflictError   = errors.New("editing Conflict")
)

type Models struct {
	Comics      ComicModel
	Tokens      TokenModel
	Users       UserModel
	Permissions PermissionModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Comics:      ComicModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}
