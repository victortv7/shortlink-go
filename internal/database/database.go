package database

import (
	"database/sql"
	"fmt"
	"log"
	"shortlink-go/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(cfg *config.Config) *sql.DB {
	dbURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %v\n", err)
	}

	log.Println("Connected to the database successfully.")
	return db
}
