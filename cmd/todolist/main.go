package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gjb1088/To-Do-list/internal/handlers"
	"github.com/gjb1088/To-Do-list/internal/models"
)

func main() {
    // ToDo store
    todoStore := models.NewStore()
    todoH, err := handlers.NewHandler(todoStore)
    if err != nil {
        log.Fatal(err)
    }

    // User store
    userStore := models.NewUserStore()
    // (Optionally seed one user)
    userStore.Create("alice", "password123")

    authH, err := handlers.NewAuthHandler(userStore)
    if err != nil {
        log.Fatal(err)
    }

    // Auth routes
    http.HandleFunc("/login", func(w, r){ 
        if r.Method=="GET"{authH.LoginPage(w,r)} else {authH.Login(w,r)} })
    http.HandleFunc("/logout", authH.Logout)

    // Protect everything else
    protected := handlers.AuthRequired(http.DefaultServeMux)

    // To-Do routes (mounted on default mux)
    http.HandleFunc("/", todoH.ServeIndex)
    http.HandleFunc("/tasks", todoH.CreateToDo)
    // ... and all your other task routes ...

    // Static files
    fs := http.FileServer(http.Dir(filepath.Join("static")))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    log.Println("Starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", protected))
}

func main() {
	// Initialize the in-memory store
	store := models.NewStore()

	// Create Handler (parses templates)
	h, err := handlers.NewHandler(store)
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	// Static assets (if any, e.g. CSS files, images)
	fs := http.FileServer(http.Dir(filepath.Join("static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", h.ServeIndex)

	// Create
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateToDo(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Get single item (for “Cancel” in inline edit, or initial load if you want to fetch a single <li>)
	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		// Decide based on method & path suffix:
		path := r.URL.Path // e.g. "/tasks/3", "/tasks/3/edit"
		switch {
		case r.Method == http.MethodGet && len(path) > len("/tasks/") && path[len(path)-len("/edit"):] == "/edit":
			h.EditFormToDo(w, r)
		case r.Method == http.MethodGet && len(path) > len("/tasks/"):
			h.GetToDo(w, r)
		case r.Method == http.MethodPut && len(path) > len("/tasks/"):
			h.UpdateToDo(w, r)
		case r.Method == http.MethodDelete && len(path) > len("/tasks/"):
			h.DeleteToDo(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/tasks/completed", func(w http.ResponseWriter, r *http.Request) {
    		if r.Method == http.MethodDelete {
        		h.ClearCompleted(w, r)
        	return
    		}
    		http.NotFound(w, r)
	})

	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
