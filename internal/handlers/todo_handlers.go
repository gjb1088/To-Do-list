package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"your_module_path/internal/models"
)

// Handler holds the store and pre-parsed templates.
type Handler struct {
	Store     *models.Store
	Templates *template.Template
}

// NewHandler parses templates and returns a Handler.
func NewHandler(store *models.Store) (*Handler, error) {
	// Parse all templates under internal/templates:
	tmpl, err := template.ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		return nil, err
	}

	// Parse partials (if you split them into a subfolder):
	partials, err := template.ParseGlob(filepath.Join("internal", "templates", "partials", "*.html"))
	if err != nil {
		return nil, err
	}

	// Combine them so that “layout.html” can reference partials:
	tmpl, err = tmpl.Clone()
	if err != nil {
		return nil, err
	}
	tmpl, err = tmpl.AddParseTree("partials", partials.Tree)
	if err != nil {
		return nil, err
	}

	return &Handler{
		Store:     store,
		Templates: tmpl,
	}, nil
}

// ServeIndex renders the full page (initial load).
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	todos := h.Store.GetAll()
	data := struct {
		ToDos []*models.ToDo
	}{
		ToDos: todos,
	}

	// layout.html should include index.html as the “main” block.
	if err := h.Templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateToDo handles POST /tasks
func (h *Handler) CreateToDo(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	title := r.PostFormValue("title")
	if title == "" {
		http.Error(w, "title cannot be empty", http.StatusBadRequest)
		return
	}

	todo := h.Store.Create(title)

	// If it’s an htmx request, we want to return just the single <li> snippet:
	if r.Header.Get("HX-Request") == "true" {
		// Render partial “todo_item.html” with this new item
		if err := h.Templates.ExecuteTemplate(w, "partials/todo_item.html", todo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Otherwise, redirect back to the index (full load)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteToDo handles DELETE /tasks/{id}
func (h *Handler) DeleteToDo(w http.ResponseWriter, r *http.Request) {
	// Assume the URL path is /tasks/{id}, e.g. /tasks/3
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Store.Delete(id); err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// For hx-delete, we return an empty 200 so htmx can remove the <li>.
	w.WriteHeader(http.StatusOK)
}

// UpdateToDo handles PUT /tasks/{id}
func (h *Handler) UpdateToDo(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	title := r.PostFormValue("title")
	completed := r.PostFormValue("completed") == "on"

	updated, err := h.Store.Update(id, title, completed)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// If this is an htmx request for inline edit, return the updated <li>
	if r.Header.Get("HX-Request") == "true" {
		if err := h.Templates.ExecuteTemplate(w, "partials/todo_item.html", updated); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditFormToDo handles GET /tasks/{id}/edit → returns an <input> form snippet
func (h *Handler) EditFormToDo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/tasks/") : len(r.URL.Path)-len("/edit")]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	todo, err := h.Store.Get(id)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// Return the snippet containing the inline edit form
	if err := h.Templates.ExecuteTemplate(w, "partials/edit_form.html", todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
