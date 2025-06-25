package models

import "github.com/jmoiron/sqlx"

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
		`SELECT id, title, completed FROM todos
		 WHERE username=$1
		 ORDER BY id`,
		username,
	)
	return todos, err
}

func (s *StorePostgres) Create(username, title string) (*ToDo, error) {
	var id int
	err := s.db.QueryRowx(
		`INSERT INTO todos(username,title)
		 VALUES($1,$2)
		 RETURNING id`,
		username, title,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &ToDo{ID: id, Title: title, Completed: false}, nil
}

func (s *StorePostgres) Update(id int, title string, completed bool, username string) (*ToDo, error) {
	_, err := s.db.Exec(
		`UPDATE todos SET title=$1, completed=$2
		 WHERE id=$3 AND username=$4`,
		title, completed, id, username,
	)
	if err != nil {
		return nil, err
	}
	return &ToDo{ID: id, Title: title, Completed: completed}, nil
}

func (s *StorePostgres) Delete(id int, username string) error {
	_, err := s.db.Exec(
		`DELETE FROM todos WHERE id=$1 AND username=$2`,
		id, username,
	)
	return err
}

func (s *StorePostgres) ClearCompleted(username string) error {
	_, err := s.db.Exec(
		`DELETE FROM todos WHERE completed=TRUE AND username=$1`,
		username,
	)
	return err
}
