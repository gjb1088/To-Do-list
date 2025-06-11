package models

import "golang.org/x/crypto/bcrypt"

// User holds a single user’s data.
type User struct {
    Username     string
    PasswordHash []byte
}

// UserStore is an in-memory user registry.
type UserStore struct {
    users map[string]*User
}

// NewUserStore initializes an empty store.
func NewUserStore() *UserStore {
    return &UserStore{users: make(map[string]*User)}
}

// Create adds a new user with a bcrypt-hashed password.
func (s *UserStore) Create(username, password string) error {
    if _, exists := s.users[username]; exists {
        return fmt.Errorf("user exists")
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    s.users[username] = &User{Username: username, PasswordHash: hash}
    return nil
}

// Authenticate returns true if username/password match.
func (s *UserStore) Authenticate(username, password string) bool {
    u, ok := s.users[username]
    if !ok {
        return false
    }
    return bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)) == nil
}
