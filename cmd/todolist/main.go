package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/gjb1088/To-Do-list/internal/handlers"
	"github.com/gjb1088/To-Do-list/internal/models"
)

func main() {
	// 1) Connect to Postgres
	db, err := sqlx.Connect(
		"pgx",
		"postgres://todoapp:secretpass@localhost:5432/todoapp?sslmode=disable",
	)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer db.Close()

	// 2) Create Postgres-backed stores
	userStore := models.NewUserStorePostgres(db)
	todoStore := models.NewStorePostgres(db)

	// 3) Build handlers
	authH, err := handlers.NewAuthHandler(userStore)
	if err != nil {
		log.Fatalf("failed to parse auth templates: %v", err)
	}
	todoH, err := handlers.NewHandlerWithStore(todoStore)
	if err != nil {
		log.Fatalf("failed to parse To-Do templates: %v", err)
	}

	// 4) Register routes on a fresh ServeMux
	mux := http.NewServeMux()

	// Unprotected auth routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authH.LoginPage(w, r)
		} else {
			authH.Login(w, r)
		}
	})
	mux.HandleFunc("/logout", authH.Logout)
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authH.RegisterPage(w, r)
		} else if r.Method == http.MethodPost {
			authH.Register(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Protected To-Do routes
	mux.Handle("/", handlers.AuthRequired(http.HandlerFunc(todoH.ServeIndex)))

	mux.Handle("/tasks", handlers.AuthRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			todoH.CreateToDo(w, r)
			return
		}
		http.NotFound(w, r)
	})))

	mux.Handle("/tasks/", handlers.AuthRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case r.Method == http.MethodGet && len(path) > len("/tasks/") && strings.HasSuffix(path, "/edit"):
			todoH.EditFormToDo(w, r)

		case r.Method == http.MethodGet && len(path) > len("/tasks/"):
			todoH.GetToDo(w, r)

		case r.Method == http.MethodPut && len(path) > len("/tasks/"):
			todoH.UpdateToDo(w, r)

		case r.Method == http.MethodDelete && len(path) > len("/tasks/"):
			todoH.DeleteToDo(w, r)

		default:
			http.NotFound(w, r)
		}
	})))

	mux.Handle("/tasks/completed", handlers.AuthRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			todoH.ClearCompleted(w, r)
			return
		}
		http.NotFound(w, r)
	})))

	// Static assets (always unprotected)
	fs := http.FileServer(http.Dir(filepath.Join("static")))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 5) Launch!
	log.Println("Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
