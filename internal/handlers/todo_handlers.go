package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gjb1088/To-Do-list/internal/models"
)

type viewData struct {
  Active    []*models.ToDo
  Completed []*models.ToDo
}

// Handler holds the store and the parsed templates.
type Handler struct {
	Store     *models.Store
	Templates *template.Template
}

// NewHandler loads all templates (layout, index, partials) and returns a Handler.
func NewHandler(store *models.Store) (*Handler, error) {
	// 1) Load layout.html and index.html
	tmpl, err := template.ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		return nil, err
	}

	// 2) Then load all the partials (todo_item.html, todo_list.html, edit_form.html)
	tmpl, err = tmpl.ParseGlob(filepath.Join("internal", "templates", "partials", "*.html"))
	if err != nil {
		return nil, err
	}

	return &Handler{
		Store:     store,
		Templates: tmpl,
	}, nil
}

// ServeIndex renders layout.html (which pulls in the "main" block from index.html)
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
    all := h.Store.GetAll()
    var active, completed []*models.ToDo
    for _, t := range all {
        if t.Completed {
            completed = append(completed, t)
        } else {
            active = append(active, t)
        }
    }

    data := struct {
        Active    []*models.ToDo
        Completed []*models.ToDo
    }{
        Active:    active,
        Completed: completed,
    }

    if err := h.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// CreateToDo handles POST /tasks
func (h *Handler) CreateToDo(w http.ResponseWriter, r *http.Request) {
  // … parse form, create todo …
  if r.Header.Get("HX-Request") == "true" {
    // build the two lists
    all := h.Store.GetAll()
    var active, completed []*models.ToDo
    for _, t := range all {
      if t.Completed {
        completed = append(completed, t)
      } else {
        active = append(active, t)
      }
    }
    data := viewData{Active: active, Completed: completed}
    // now re-render the entire #todoApp
    if err := h.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    return
  }
  http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteToDo handles DELETE /tasks/{id}
func (h *Handler) DeleteToDo(w http.ResponseWriter, r *http.Request) {
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

	// htmx will remove the <li> if we return 200 OK with empty body
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

	// For htmx inline updates, return the new <li>
	if r.Header.Get("HX-Request") == "true" {
		if err := h.Templates.ExecuteTemplate(w, "todo_item.html", updated); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditFormToDo handles GET /tasks/{id}/edit and returns the inline edit form <li>
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

	if err := h.Templates.ExecuteTemplate(w, "edit_form.html", todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetToDo handles GET /tasks/{id} and returns a single <li> for htmx swaps
func (h *Handler) GetToDo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	todo, err := h.Store.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Return just the <li> snippet
	if r.Header.Get("HX-Request") == "true" {
		if err := h.Templates.ExecuteTemplate(w, "todo_item.html", todo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ClearCompleted handles DELETE /tasks/completed
func (h *Handler) ClearCompleted(w http.ResponseWriter, r *http.Request) {
    h.Store.ClearCompleted()
    // Return an empty <ul> so htmx will wipe out the list
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(`<ul id="completedList"></ul>`))
}
