package models

// UserStore abstracts how we create/authenticate users.
type UserStore interface {
    Create(username, password string) error
    Authenticate(username, password string) bool
}

// ToDoStore abstracts how we CRUD todos for a given user.
type ToDoStore interface {
    // List all to-dos for this username
    GetAll(username string) ([]*ToDo, error)
    // Create a new to-do, return the record
    Create(username, title string) (*ToDo, error)
    // Update title/completed for a given ID (and username)
    Update(id int, title string, completed bool, username string) (*ToDo, error)
    // Delete by ID (and username)
    Delete(id int, username string) error
    // Remove all completed for this username
    ClearCompleted(username string) error
}
