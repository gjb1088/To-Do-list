package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gjb1088/To-Do-list/internal/models"
)

// pageData holds everything layout.html needs: the current user + both lists.
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

// Handler bundles your ToDoStore and parsed templates.
type Handler struct {
	store     models.ToDoStore
	Templates *template.Template
}

// NewHandlerWithStore parses your layout + all partials and returns a Handler
// wired to any models.ToDoStore implementation.
func NewHandlerWithStore(store models.ToDoStore) (*Handler, error) {
	// 1) Parse layout.html + index.html
	tmpl, err := template.ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		return nil, err
	}
	// 2) Parse all partials: todo_item.html, todo_list.html, edit_form.html, etc.
	tmpl, err = tmpl.ParseGlob(filepath.Join("internal", "templates", "partials", "*.html"))
	if err != nil {
		return nil, err
	}
	return &Handler{store: store, Templates: tmpl}, nil
}

// currentUser pulls the signed-in username out of the session cookie.
func (h *Handler) currentUser(r *http.Request) string {
	sess, _ := sessionStore.Get(r, sessionName)
	user, _ := sess.Values["user"].(string)
	return user
}

// buildViewData fetches all todos for this user and splits into active/completed.
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

// ServeIndex handles GET "/" and renders the full page (using layout.html).
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	user := h.currentUser(r)
	vd := h.buildViewData(user)
	data := pageData{
		Username:  user,
		Active:    vd.Active,
		Completed: vd.Completed,
	}
	if err := h.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateToDo handles POST "/tasks". On HTMX it re-renders only the "main" block.
func (h *Handler) CreateToDo(w http.ResponseWriter, r *http.Request) {
	// 1) Parse + validate
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	title := r.PostFormValue("title")
	if title == "" {
		http.Error(w, "title cannot be empty", http.StatusBadRequest)
		return
	}

	// 2) Create under the current user
	user := h.currentUser(r)
	if _, err := h.store.Create(user, title); err != nil {
		http.Error(w, "could not create todo", http.StatusInternalServerError)
		return
	}

	// 3) If HTMX, re-render the <div id="todoApp">…</div> by firing the "main" template
	if r.Header.Get("HX-Request") == "true" {
		vd := h.buildViewData(user)
		data := pageData{
			Username:  user,
			Active:    vd.Active,
			Completed: vd.Completed,
		}
		if err := h.Templates.ExecuteTemplate(w, "main", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 4) Fallback: full redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteToDo handles DELETE "/tasks/{id}" and simply returns 200 OK on HTMX so
// htmx will remove the <li> for you.
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

// UpdateToDo handles PUT "/tasks/{id}" – distinguishes inline edit vs checkbox toggle.
func (h *Handler) UpdateToDo(w http.ResponseWriter, r *http.Request) {
	// 1) Parse the form
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

	// 2) Load the old todo (for title fallback)
	old, err := h.store.Get(id, user)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	// 3) Determine new values
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

	// 4) HTMX inline-edit vs toggle:
	if r.Header.Get("HX-Request") == "true" {
		// a) inline save → return a single <li> snippet
		if r.PostFormValue("title") != "" {
			if err := h.Templates.ExecuteTemplate(w, "partials/todo_item.html", updated); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// b) checkbox toggle → re-render entire todoApp
		vd := h.buildViewData(user)
		data := pageData{
			Username:  user,
			Active:    vd.Active,
			Completed: vd.Completed,
		}
		if err := h.Templates.ExecuteTemplate(w, "main", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 5) non-HTMX fallback
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditFormToDo handles GET "/tasks/{id}/edit" → returns the inline edit <li> form.
func (h *Handler) EditFormToDo(w http.ResponseWriter, r *http.Request) {
	// chop off "/tasks/" … "/edit"
	raw := r.URL.Path[len("/tasks/") : len(r.URL.Path)-len("/edit")]
	id, err := strconv.Atoi(raw)
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

// GetToDo handles GET "/tasks/{id}" → returns a single <li> snippet.
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

// ClearCompleted handles DELETE "/tasks/completed" → re-renders the main block.
func (h *Handler) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	user := h.currentUser(r)
	h.store.ClearCompleted(user)

	vd := h.buildViewData(user)
	data := pageData{
		Username:  user,
		Active:    vd.Active,
		Completed: vd.Completed,
	}
	h.Templates.ExecuteTemplate(w, "main", data)
}
