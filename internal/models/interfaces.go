package models

// UserStore abstracts how we create/authenticate users.
type UserStore interface {
    Create(username, password string) error
    Authenticate(username, password string) bool
}

// ToDoStore abstracts how we CRUD todos for a given user.
type ToDoStore interface {
    // fetch all to-dos for this user
    GetAll(username string) ([]*ToDo, error)
    // fetch one by ID+user
    Get(id int, username string) (*ToDo, error)
    Create(username, title string) (*ToDo, error)
    Update(id int, title string, completed bool, username string) (*ToDo, error)
    Delete(id int, username string) error
    ClearCompleted(username string) error
}
