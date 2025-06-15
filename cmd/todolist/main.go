package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gjb1088/To-Do-list/internal/handlers"
	"github.com/gjb1088/To-Do-list/internal/models"
)

func main() {
	// --- 1) Set up the To-Do service ---

	// In-memory To-Do store + handler
	todoStore := models.NewStore()
	todoH, err := handlers.NewHandler(todoStore)
	if err != nil {
		log.Fatalf("failed to parse To-Do templates: %v", err)
	}

	// --- 2) Set up the Auth service ---

	userStore := models.NewUserStore()
	// Seed a default user
	if err := userStore.Create("alice", "password123"); err != nil {
		log.Fatalf("failed to seed user: %v", err)
	}
	authH, err := handlers.NewAuthHandler(userStore)
	if err != nil {
		log.Fatalf("failed to parse auth templates: %v", err)
	}

	// --- 3) Build a new ServeMux and register all routes ---

	mux := http.NewServeMux()

	// Auth routes (unprotected)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authH.LoginPage(w, r)
		} else {
			authH.Login(w, r)
		}
	})
	mux.HandleFunc("/logout", authH.Logout)

	// To-Do routes (all protected below)
	mux.HandleFunc("/", todoH.ServeIndex)
	mux.HandleFunc("/tasks", todoH.CreateToDo)
	// Single-item routes: GET /tasks/{id}, PUT, DELETE, and /tasks/{id}/edit
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path // e.g. "/tasks/3", "/tasks/3/edit"
		switch {
		case r.Method == http.MethodPost:
			http.NotFound(w, r) // POST /tasks goes to /tasks above
		case r.Method == http.MethodGet && len(path) > len("/tasks/") && path[len(path)-len("/edit"):] == "/edit":
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
	})
	// Bulk-clear completed
	mux.HandleFunc("/tasks/completed", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			todoH.ClearCompleted(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Static assets
	fs := http.FileServer(http.Dir(filepath.Join("static")))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// --- 4) Wrap the mux in AuthRequired, so only /login & /logout work for anonymous---

	protectedMux := handlers.AuthRequired(mux)

	// --- 5) Start the server ---
	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", protectedMux); err != nil {
		log.Fatal(err)
	}
}
