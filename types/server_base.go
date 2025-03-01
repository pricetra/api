package types

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/go-playground/validator/v10"
)

type ServerBase struct {
	DB *sql.DB
	Router *chi.Mux
	StructValidator *validator.Validate
	MigrationDirectory string
	Port string
	Tokens *Tokens
}
