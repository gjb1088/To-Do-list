package handlers

import (
    "html/template"
    "net/http"

    "github.com/gorilla/sessions"
    "github.com/gjb1088/To-Do-list/internal/models"
)

var (
    // sessionStore holds our cookie-based session backend.
    sessionStore = sessions.NewCookieStore([]byte("super-secret-key"))
    sessionName  = "todo-session"
)

// AuthHandler bundles a UserStore (interface) + templates.
type AuthHandler struct {
    userStore models.UserStore
    Templates *template.Template
}

// NewAuthHandler parses the auth templates and wires in any UserStore.
func NewAuthHandler(us models.UserStore) (*AuthHandler, error) {
    tmpl, err := template.ParseGlob("internal/templates/auth/*.html")
    if err != nil {
        return nil, err
    }
    return &AuthHandler{
        userStore: us,       // ← this must match the struct field
        Templates: tmpl,
    }, nil
}

// LoginPage shows the GET /login form.
func (a *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
    a.Templates.ExecuteTemplate(w, "login.html", nil)
}

// Login POSTs /login, checks credentials, and sets the session.
func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "invalid form", http.StatusBadRequest)
        return
    }
    user := r.FormValue("username")
    pass := r.FormValue("password")

    // authenticate against our store
    if a.userStore.Authenticate(user, pass) {
        sess, _ := sessionStore.Get(r, sessionName)
        sess.Values["user"] = user
        sess.Save(r, w)
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    http.Error(w, "Invalid credentials", http.StatusForbidden)
}

// Logout GETs /logout and clears the session.
func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    sess, _ := sessionStore.Get(r, sessionName)
    delete(sess.Values, "user")
    sess.Save(r, w)
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// RegisterPage shows the GET /register form.
func (a *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
    a.Templates.ExecuteTemplate(w, "register.html", nil)
}

// Register POSTs /register and creates a new user.
func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "invalid form", http.StatusBadRequest)
        return
    }
    username := r.FormValue("username")
    password := r.FormValue("password")
    if username == "" || password == "" {
        http.Error(w, "username & password required", http.StatusBadRequest)
        return
    }

    // create via our interface
    if err := a.userStore.Create(username, password); err != nil {
        http.Error(w, "user already exists", http.StatusConflict)
        return
    }

    // on success, send them to login
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AuthRequired is middleware that redirects anonymous users to /login.
func AuthRequired(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sess, _ := sessionStore.Get(r, sessionName)
        if _, ok := sess.Values["user"]; !ok {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        next.ServeHTTP(w, r)
    })
}
