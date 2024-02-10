package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/config"

	// _ "github.com/denisenkom/go-mssqldb"
	_ "github.com/lib/pq"
)

func Initialize(settings *config.Config) (*sqlx.DB, error) {

	var db *sqlx.DB

	// Construct connection string

	// MSSQL Server specific
	// connString := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;",
	// 	settings.DbHost, settings.DbPort, settings.DbUser, settings.DbPassword, settings.DbName)
	// db, err := sqlx.Open("sqlserver", connString)
	// POSTGRESQL specific
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", settings.DbHost, settings.DbPort, settings.DbUser, settings.DbPassword, settings.DbName)
	db, err := sqlx.Open("postgres", connString)

	if err != nil {
		log.Err(err).Msg("Error initializing database.")
		return nil, err
	}

	// test the connection
	err = db.Ping()

	if err != nil {
		log.Err(err).Msg("Error pinging the database.")
		return nil, err
	}

	log.Info().Msg("Database Initialized Successfully.")
	return db, nil
}
