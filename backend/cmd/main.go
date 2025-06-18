package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
)

var db *sql.DB

func init() {
	var err error
	const connection = "postgres://radat:radatSWP25@postgres:5432/radatDB?sslmode=disable"
	db, err = sql.Open("postgres", connection)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	runMigrations()
}

func runMigrations() {
	migrationPath := "file:///migrations"
	const connection = "postgres://radat:radatSWP25@postgres:5432/radatDB?sslmode=disable"
	m, err := migrate.New(migrationPath, connection)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Error applying migrations: ", err)
	}
	log.Println("Migrations have been successfully applied!")
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("Добро пожаловать на главную страницу!"))
	})
	http.HandleFunc("/api/start", startHandler)
	http.HandleFunc("/api/message", messageHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
