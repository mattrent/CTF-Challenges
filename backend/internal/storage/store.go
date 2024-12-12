package storage

import (
	"database/sql"
	"deployer/config"
	"deployer/internal/auth"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var Db *sql.DB

func InitDb() {
	var err error
	Db, err = sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		log.Println("Opening connection to DB")
		log.Fatal(err.Error())
	}

	err = Db.Ping()
	if err != nil {
		log.Println("Testing connection to DB")
		log.Fatal(err.Error())
	}

	log.Println("Running migrations")
	m, err := migrate.New(
		"file://migrations",
		config.Values.DbConn)
	if err != nil {
		log.Fatal(err.Error())
	}
	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		log.Println(err)
	} else if err != nil {
		log.Fatalln(err)
	}

	// Seed with initial user
	if len(config.Values.AdminPassword) > 0 {
		hashedPassword, err := auth.HashPassword(config.Values.AdminPassword)
		if err != nil {
			log.Fatal(err.Error())
		}
		res, err := Db.Exec("INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) ON CONFLICT (username) DO NOTHING", "admin", hashedPassword, "admin")
		if err != nil {
			log.Fatal(err.Error())
		}
		rowCount, err := res.RowsAffected()
		if err != nil {
			log.Fatal(err.Error())
		}
		if rowCount > 0 {
			log.Println("Created admin user")
		}
	}
}
