package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gjb1088/To-Do-list/internal/models"
)

// pageData holds everything layout.html expects: the current user and both lists.
type pageData struct {
	Username  string
	Active    []*models.ToDo
	Completed []*models.ToDo
}

// viewData just holds the two To-Do slices.
type viewData struct {
	Active    []*models.ToDo
	Completed []*models.ToDo
}

// Handler bundles the To-Do store and parsed templates.
type Handler struct {
	Store     *models.Store
	Templates *template.Template
}

// NewHandler parses layout/index + all partials into one *template.Template.
func NewHandler(store *models.Store) (*Handler, error) {
	tmpl, err := template.ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		return nil, err
	}
	tmpl, err = tmpl.ParseGlob(filepath.Join("internal", "templates", "partials", "*.html"))
	if err != nil {
		return nil, err
	}
	return &Handler{Store: store, Templates: tmpl}, nil
}

// buildViewData splits all To-Dos into Active vs Completed.
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

// ServeIndex handles GET "/" – pulls the username from the session + both lists,
// then renders layout.html (which calls your main block).
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	// 1) Get the signed-in username from the session.
	sess, _ := sessionStore.Get(r, sessionName)
	username, _ := sess.Values["user"].(string)

	// 2) Build the two lists.
	vd := h.buildViewData()

	// 3) Combine into the full pageData.
	data := pageData{
		Username:  username,
		Active:    vd.Active,
		Completed: vd.Completed,
	}

	// 4) Render the full shell (layout.html with its main block).
	if err := h.Templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateToDo handles POST /tasks. If htmx, re-renders only the main block.
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

	if r.Header.Get("HX-Request") == "true" {
		data := h.composePageData(r)
		h.Templates.ExecuteTemplate(w, "main", data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DeleteToDo handles DELETE /tasks/{id}. htmx will remove the <li> via outerHTML.
func (h *Handler) DeleteToDo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
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

// UpdateToDo handles PUT /tasks/{id}. Distinguishes inline-edit vs checkbox toggle.
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
	// Fetch old for title fallback
	old, err := h.Store.Get(id)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	// Title from form if present, else old.Title
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

	if r.Header.Get("HX-Request") == "true" {
		// Inline-edit (title present) → single <li>
		if r.PostFormValue("title") != "" {
			h.Templates.ExecuteTemplate(w, "todo_item.html", updated)
			return
		}
		// Checkbox toggle → re-render main block
		data := h.composePageData(r)
		h.Templates.ExecuteTemplate(w, "main", data)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditFormToDo handles GET /tasks/{id}/edit → returns the edit_form.html snippet.
func (h *Handler) EditFormToDo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/") : len(r.URL.Path)-len("/edit")])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	todo, err := h.Store.Get(id)
	if err != nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}
	h.Templates.ExecuteTemplate(w, "edit_form.html", todo)
}

// GetToDo handles GET /tasks/{id} → returns the todo_item.html snippet.
func (h *Handler) GetToDo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	todo, err := h.Store.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	h.Templates.ExecuteTemplate(w, "todo_item.html", todo)
}

// ClearCompleted handles DELETE /tasks/completed → re-renders only the main block.
func (h *Handler) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	h.Store.ClearCompleted()
	data := h.composePageData(r)
	h.Templates.ExecuteTemplate(w, "main", data)
}

// composePageData is an internal helper to pull username + lists for htmx swaps.
func (h *Handler) composePageData(r *http.Request) pageData {
	// session + username
	sess, _ := sessionStore.Get(r, sessionName)
	username, _ := sess.Values["user"].(string)
	vd := h.buildViewData()
	return pageData{
		Username:  username,
		Active:    vd.Active,
		Completed: vd.Completed,
	}
}
