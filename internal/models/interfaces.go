package models

import (
    "errors"
    "time"
)	

// ToDo is your database model for a single task.
type ToDo struct {
	ID        int       `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Completed bool      `db:"completed" json:"completed"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ErrNotFound is returned when a to-do item doesn’t exist.
var ErrNotFound = errors.New("todo not found")

// UserStore abstracts how we create/authenticate users.
type UserStore interface {
	Create(username, password string) error
	Authenticate(username, password string) bool
}

// ToDoStore abstracts how we CRUD todos for a given user.
type ToDoStore interface {
	// Fetch all to-dos for this user.
	GetAll(username string) ([]*ToDo, error)
	// Fetch one to-do by ID and user.
	Get(id int, username string) (*ToDo, error)
	// Create a new to-do.
	Create(username, title string) (*ToDo, error)
	// Update title/completed flag.
	Update(id int, title string, completed bool, username string) (*ToDo, error)
	// Delete one to-do.
	Delete(id int, username string) error
	// Remove all completed items.
	ClearCompleted(username string) error
}
