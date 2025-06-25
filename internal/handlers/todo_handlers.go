package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gjb1088/To-Do-list/internal/models"
)

// pageData holds everything layout.html expects.
type pageData struct {
	Username  string
	Active    []*models.ToDo
	Completed []*models.ToDo
}

// viewData holds just the two To-Do slices.
type viewData struct {
	Active    []*models.ToDo
	Completed []*models.ToDo
}

// Handler bundles our ToDoStore interface and parsed templates.
type Handler struct {
	store     models.ToDoStore
	Templates *template.Template
}

// NewHandlerWithStore parses layout + partials and returns a handler
// wired to *any* models.ToDoStore implementation.
func NewHandlerWithStore(store models.ToDoStore) (*Handler, error) {
	tmpl, err := template.ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		return nil, err
	}
	tmpl, err = tmpl.ParseGlob(filepath.Join("internal", "templates", "partials", "*.html"))
	if err != nil {
		return nil, err
	}
	return &Handler{store: store, Templates: tmpl}, nil
}

// buildViewData fetches *all* the user's todos and splits them.
func (h *Handler) buildViewData(username string) viewData {
	all, _ := h.store.GetAll(username)
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

// ServeIndex GET "/" → renders the full page (layout.html).
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	user := h.currentUser(r)
	data := pageData{
		Username:  user,
		Active:    h.buildViewData(user).Active,
		Completed: h.buildViewData(user).Completed,
	}
	if err := h.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateToDo POST "/tasks" → adds a new to-do for this user.
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

	user := h.currentUser(r)
	h.store.Create(user, title)

	// If htmx, swap in the updated main block:
	if r.Header.Get("HX-Request") == "true" {
		data := pageData{Username: user}
		vd := h.buildViewData(user)
		data.Active, data.Completed = vd.Active, vd.Completed
		h.Templates.ExecuteTemplate(w, "main", data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteToDo DELETE "/tasks/{id}" → removes that item for this user.
func (h *Handler) DeleteToDo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	user := h.currentUser(r)
	if err := h.store.Delete(id, user); err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateToDo PUT "/tasks/{id}" → updates title/completed for this user.
func (h *Handler) UpdateToDo(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	user := h.currentUser(r)

	// Fetch old for title fallback
	old, _ := h.store.Get(id, user)
	title := r.PostFormValue("title")
	if title == "" {
		title = old.Title
	}
	completed := r.PostFormValue("completed") == "on"

	updated, err := h.store.Update(id, title, completed, user)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		// Inline edit (title present) → single <li>
		if r.PostFormValue("title") != "" {
			h.Templates.ExecuteTemplate(w, "partials/todo_item.html", updated)
			return
		}
		// Checkbox toggle → re-render main block
		data := pageData{Username: user}
		vd := h.buildViewData(user)
		data.Active, data.Completed = vd.Active, vd.Completed
		h.Templates.ExecuteTemplate(w, "main", data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditFormToDo GET "/tasks/{id}/edit" → returns the <li> edit form snippet.
func (h *Handler) EditFormToDo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(
		r.URL.Path[len("/tasks/") : len(r.URL.Path)-len("/edit")],
	)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	user := h.currentUser(r)
	todo, err := h.store.Get(id, user)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	h.Templates.ExecuteTemplate(w, "partials/edit_form.html", todo)
}

// GetToDo GET "/tasks/{id}" → returns the single <li> snippet.
func (h *Handler) GetToDo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	user := h.currentUser(r)
	todo, err := h.store.Get(id, user)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	h.Templates.ExecuteTemplate(w, "partials/todo_item.html", todo)
}

// ClearCompleted DELETE "/tasks/completed" → re-renders main block.
func (h *Handler) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	user := h.currentUser(r)
	h.store.ClearCompleted(user)

	// Return updated main block
	data := pageData{Username: user}
	vd := h.buildViewData(user)
	data.Active, data.Completed = vd.Active, vd.Completed
	h.Templates.ExecuteTemplate(w, "main", data)
}

// currentUser is a helper to pull "user" out of the session cookie.
func (h *Handler) currentUser(r *http.Request) string {
	sess, _ := sessionStore.Get(r, sessionName)
	user, _ := sess.Values["user"].(string)
	return user
}
