package models

import (
    "github.com/jmoiron/sqlx"
)

// StorePostgres implements ToDoStore for a PostgreSQL backend.
type StorePostgres struct {
    db *sqlx.DB
}

func NewStorePostgres(db *sqlx.DB) *StorePostgres {
    return &StorePostgres{db: db}
}

func (s *StorePostgres) GetAll(username string) ([]*ToDo, error) {
    var todos []*ToDo
    err := s.db.Select(
        &todos,
        `SELECT id, title, completed, created_at, updated_at
           FROM todos
          WHERE username = $1
          ORDER BY id`,
        username,
    )
    return todos, err
}

func (s *StorePostgres) Get(id int, username string) (*ToDo, error) {
    var todo ToDo
    err := s.db.Get(
        &todo,
        `SELECT id, title, completed, created_at, updated_at
           FROM todos
          WHERE id = $1 AND username = $2`,
        id, username,
    )
    if err != nil {
        return nil, err
    }
    return &todo, nil
}

func (s *StorePostgres) Create(username, title string) (*ToDo, error) {
    var t ToDo
    err := s.db.Get(
        &t,
        `INSERT INTO todos (username, title)
             VALUES ($1, $2)
         RETURNING id, title, completed, created_at, updated_at`,
        username, title,
    )
    if err != nil {
        return nil, err
    }
    return &t, nil
}

func (s *StorePostgres) Update(
    id int, title string, completed bool, username string,
) (*ToDo, error) {
    var t ToDo
    err := s.db.Get(
        &t,
        `UPDATE todos
            SET title     = $1,
                completed = $2,
                updated_at = NOW()
          WHERE id       = $3
            AND username = $4
      RETURNING id, title, completed, created_at, updated_at`,
        title, completed, id, username,
    )
    if err != nil {
        return nil, err
    }
    return &t, nil
}

func (s *StorePostgres) Delete(id int, username string) error {
    _, err := s.db.Exec(
        `DELETE FROM todos
          WHERE id       = $1
            AND username = $2`,
        id, username,
    )
    return err
}

func (s *StorePostgres) ClearCompleted(username string) error {
    _, err := s.db.Exec(
        `DELETE FROM todos
          WHERE completed = TRUE
            AND username  = $1`,
        username,
    )
    return err
}
