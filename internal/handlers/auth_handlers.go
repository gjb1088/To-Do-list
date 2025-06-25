package handlers

import (
    "html/template"
    "net/http"

    "github.com/gorilla/sessions"
    "github.com/gjb1088/To-Do-list/internal/models"
)

var (
    // Replace with a strong secret in env/config
    sessionStore = sessions.NewCookieStore([]byte("super-secret-key"))
    sessionName  = "todo-session"
)

// AuthHandler holds user store + templates
type AuthHandler struct {
    Users     models.UserStore
    Templates *template.Template
}

// NewAuthHandler loads auth templates (login.html, register.html)
func NewAuthHandler(us models.UserStore) (*AuthHandler, error) {
    tmpl, err := template.ParseGlob("internal/templates/auth/*.html")
    if err != nil {
        return nil, err
    }
    return &AuthHandler{userStore: us, Templates: tmpl}, nil
}

// LoginPage GET /login
func (a *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
    a.Templates.ExecuteTemplate(w, "login.html", nil)
}

// Login POST /login
func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    user := r.FormValue("username")
    pass := r.FormValue("password")
    if a.userStore.Authenticate(user, pass) {
        sess, _ := sessionStore.Get(r, sessionName)
        sess.Values["user"] = user
        sess.Save(r, w)
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    http.Error(w, "Invalid credentials", http.StatusForbidden)
}

// Logout GET /logout
func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    sess, _ := sessionStore.Get(r, sessionName)
    delete(sess.Values, "user")
    sess.Save(r, w)
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AuthRequired wraps a handler, redirecting to /login if not signed in.
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

// RegisterPage handles GET /register and shows the register form.
func (a *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
    a.Templates.ExecuteTemplate(w, "register.html", nil)
}

// Register handles POST /register and creates a new user.
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
    if err := a.userStore.Create(username, password); err != nil {
        // you might want to show a nicer error page instead
        http.Error(w, "user already exists", http.StatusConflict)
        return
    }
    // After successful registration, redirect to login
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}
