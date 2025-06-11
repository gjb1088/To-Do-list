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
    Users     *models.UserStore
    Templates *template.Template
}

// NewAuthHandler loads auth templates (login.html, register.html)
func NewAuthHandler(users *models.UserStore) (*AuthHandler, error) {
    tmpl, err := template.ParseGlob("internal/templates/auth/*.html")
    if err != nil {
        return nil, err
    }
    return &AuthHandler{Users: users, Templates: tmpl}, nil
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
    if a.Users.Authenticate(user, pass) {
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
