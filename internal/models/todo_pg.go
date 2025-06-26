package models

import (
	"github.com/jmoiron/sqlx"
)

// StorePostgres implements ToDoStore for a PostgreSQL backend.
type StorePostgres struct {
	db *sqlx.DB
}

// NewStorePostgres returns a ToDoStore backed by the given sqlx.DB.
func NewStorePostgres(db *sqlx.DB) *StorePostgres {
	return &StorePostgres{db: db}
}

// GetAll returns every to‐do belonging to the given user.
func (s *StorePostgres) GetAll(username string) ([]*ToDo, error) {
	var todos []*ToDo
	err := s.db.Select(
		&todos,
		`SELECT id, title, completed
		   FROM todos
		  WHERE username = $1
		  ORDER BY id`,
		username,
	)
	return todos, err
}

// Get fetches a single to‐do by ID and username.
func (s *StorePostgres) Get(id int, username string) (*ToDo, error) {
	var todo ToDo
	err := s.db.Get(
		&todo,
		`SELECT id, title, completed
		   FROM todos
		  WHERE id = $1
		    AND username = $2`,
		id, username,
	)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

// Create inserts a new to‐do for this user and returns the created record.
func (s *StorePostgres) Create(username, title string) (*ToDo, error) {
	var id int
	err := s.db.QueryRowx(
		`INSERT INTO todos (username, title)
		     VALUES ($1, $2)
		  RETURNING id`,
		username, title,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &ToDo{ID: id, Title: title, Completed: false}, nil
}

// Update changes the title and/or completed‐flag of an existing to‐do.
func (s *StorePostgres) Update(id int, title string, completed bool, username string) (*ToDo, error) {
	_, err := s.db.Exec(
		`UPDATE todos
		    SET title     = $1,
		        completed = $2
		  WHERE id       = $3
		    AND username = $4`,
		title, completed, id, username,
	)
	if err != nil {
		return nil, err
	}
	return &ToDo{ID: id, Title: title, Completed: completed}, nil
}

// Delete removes a to‐do by ID (scoped to the signed‐in user).
func (s *StorePostgres) Delete(id int, username string) error {
	_, err := s.db.Exec(
		`DELETE FROM todos
		  WHERE id       = $1
		    AND username = $2`,
		id, username,
	)
	return err
}

// ClearCompleted deletes all “completed = TRUE” rows for that user.
func (s *StorePostgres) ClearCompleted(username string) error {
	_, err := s.db.Exec(
		`DELETE FROM todos
		  WHERE completed = TRUE
		    AND username  = $1`,
		username,
	)
	return err
}
