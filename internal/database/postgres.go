package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgres(connString string) *sql.DB {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return db
}
