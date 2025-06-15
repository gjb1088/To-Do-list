package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gjb1088/To-Do-list/internal/models"
)

)

type pageData struct {
    Username  string
    Active    []*models.ToDo
    Completed []*models.ToDo
}

func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
    // pull username from the session
    sess, _ := sessionStore.Get(r, sessionName)
    username, _ := sess.Values["user"].(string)

    // build active/completed lists
    vd := h.buildViewData()

    // >>> Correct composite literal syntax: each field as Name: value, <<<
    data := pageData{
        Username:  username,
        Active:    vd.Active,
        Completed: vd.Completed,
    }

    // render your layout
    if err := h.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// Handler bundles our store & templates.
type Handler struct {
	Store     *models.Store
	Templates *template.Template
}

// NewHandler parses layout.html, index.html, and all partials into one Template.
func NewHandler(store *models.Store) (*Handler, error) {
	// 1) parse layout.html + index.html
	tmpl, err := template.ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		return nil, err
	}
	// 2) parse all partials (todo_item.html, todo_list.html, edit_form.html, etc.)
	tmpl, err = tmpl.ParseGlob(filepath.Join("internal", "templates", "partials", "*.html"))
	if err != nil {
		return nil, err
	}
	return &Handler{Store: store, Templates: tmpl}, nil
}

// buildViewData splits todos into Active vs Completed.
func (h *Handler) buildViewData() viewData {
	all := h.Store.GetAll()
	var active, completed []*models.ToDo
	for _, t := range all {
		if t.Completed {
			completed = append(completed, t)
		} else {
			active = append(active, t)
		}
	}
	return viewData{Active: active, Completed: completed}
}

// right above your methods
type pageData struct {
    Username  string
    Active    []*models.ToDo
    Completed []*models.ToDo
}

// ServeIndex handles the initial GET "/" and renders the full layout (with one header).
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
    // 1) Grab the session and pull out the username (set in auth_handlers)
    sess, _ := sessionStore.Get(r, sessionName)
    username, _ := sess.Values["user"].(string)

    // 2) Build your active/completed slices
    vd := h.buildViewData()

    // 3) Combine username + lists into one struct for the template
    data := pageData {
        Username  username
        Active    vd.Active
        Completed vd.Completed,
    }

    // 4) Render the shell + inner block
    if err := h.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// CreateToDo handles POST /tasks. On htmx requests, re-renders only the inner "main" block.
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
	h.Store.Create(title)

	// If htmx, swap in only the updated main block.
	if r.Header.Get("HX-Request") == "true" {
		data := h.buildViewData()
		if err := h.Templates.ExecuteTemplate(w, "main", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	// Otherwise fallback to full reload.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteToDo handles DELETE /tasks/{id}. htmx will remove the <li> via outerHTML.
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
	w.WriteHeader(http.StatusOK)
}

// UpdateToDo handles PUT /tasks/{id} (toggle or inline edit)
func (h *Handler) UpdateToDo(w http.ResponseWriter, r *http.Request) {
    // Parse form so we can see “completed” and “title” if present
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

    // Fetch the old ToDo so we can fall back to its title if none is provided
    old, err := h.Store.Get(id)
    if err != nil {
        http.Error(w, "todo not found", http.StatusNotFound)
        return
    }

    // Determine new title: if inline‐edit provided one, use it; otherwise keep old
    title := r.PostFormValue("title")
    if title == "" {
        title = old.Title
    }
    completed := r.PostFormValue("completed") == "on"

    updated, err := h.Store.Update(id, title, completed)
    if err != nil {
        http.Error(w, "todo not found", http.StatusNotFound)
        return
    }

    // Handle htmx requests specially
    if r.Header.Get("HX-Request") == "true" {
        // Inline‐edit save (title was provided) → return single <li> partial
        if r.PostFormValue("title") != "" {
            if err := h.Templates.ExecuteTemplate(w, "todo_item.html", updated); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
            return
        }
        // Checkbox toggle → re-render the entire main block
        data := h.buildViewData()
        if err := h.Templates.ExecuteTemplate(w, "main", data); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    // Non-htmx fallback: full-page redirect
    http.Redirect(w, r, "/", http.StatusSeeOther)
}


// EditFormToDo handles GET /tasks/{id}/edit and returns the edit form <li>.
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
	// Return only the edit_form.html snippet.
	if err := h.Templates.ExecuteTemplate(w, "edit_form.html", todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetToDo handles GET /tasks/{id} and returns the single <li> snippet.
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
	if err := h.Templates.ExecuteTemplate(w, "todo_item.html", todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ClearCompleted handles DELETE /tasks/completed and re-renders only the main block.
func (h *Handler) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	h.Store.ClearCompleted()
	data := h.buildViewData()
	if err := h.Templates.ExecuteTemplate(w, "main", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
