package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gjb1088/To-Do-list/internal/handlers"
	"github.com/gjb1088/To-Do-list/internal/models"
)

func main() {
	// --- 1) Set up stores & handlers ---

	// To-Do store & handler
	todoStore := models.NewStore()
	todoH, err := handlers.NewHandler(todoStore)
	if err != nil {
		log.Fatalf("failed to parse To-Do templates: %v", err)
	}

	// User store & handler
	userStore := models.NewUserStore()
	// seed a user
	if err := userStore.Create("alice", "password123"); err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	authH, err := handlers.NewAuthHandler(userStore)
	if err != nil {
		log.Fatalf("failed to parse auth templates: %v", err)
	}

	// --- 2) Build a new ServeMux and register routes ---

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
		path := r.URL.Path // e.g. "/tasks/3", "/tasks/3/edit"
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

	// --- 3) Start the server ---

	log.Println("Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
