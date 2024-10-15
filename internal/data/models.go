package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Videos     VideoModel
	Categories CategoryModel
	Actors     ActorModel
	Documents  DocumentModel
	Users      UserModel
	Tokens     TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Videos:     VideoModel{DB: db},
		Categories: CategoryModel{DB: db},
		Actors:     ActorModel{DB: db},
		Documents:  DocumentModel{DB: db},
		Users:      UserModel{DB: db},
		Tokens:     TokenModel{DB: db},
	}
}
