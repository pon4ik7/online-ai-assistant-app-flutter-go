package main

import (
	"database/sql"
	"encoding/json"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5"
	"log"
	"net/http"
)

var db *sql.DB

func init() {
	var err error
	const connection = "postgres://radat:radatSWP25@localhost:5432/radatDB?sslmode=disable"
	db, err = sql.Open("postgres", connection)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	runMigrations()
}

func runMigrations() {
	migrationPath := "file://migrations"
	const connection = "postgres://radat:radatSWP25@localhost:5432/radatDB?sslmode=disable"
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
	http.HandleFunc("/api/products/add", addProduct)
	http.HandleFunc("/api/sessions/create", createSession)
	http.HandleFunc("/api/sessions/messages/add", addMessage)

	log.Println("Сервер запущен на :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// Several basic DB-requests for testing and future development

func addProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := db.Exec(`
        INSERT INTO popular_products 
        (name, description, price, rating, category, product_url, image_url) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		"Новый продукт", "Описание продукта", 100.00, 4.5, "Категория", "http://example.com", "http://example.com/image.jpg")
	if err != nil {
		http.Error(w, "Failed to add product: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Product added successfully"))
}

func createSession(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New()
	_, err := db.Exec(`
        INSERT INTO sessions 
        (session_id, context) 
        VALUES ($1, $2)`,
		sessionID, "Контекст сессии")
	if err != nil {
		http.Error(w, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID.String(),
		"status":     "created",
	})
}

func addMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	_, err := db.Exec(`
        INSERT INTO session_messages 
        (session_id, message, response) 
        VALUES ($1, $2, $3)`,
		sessionID, "Вопрос пользователя", "Ответ системы")
	if err != nil {
		http.Error(w, "Failed to add message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Message added successfully"))
}
