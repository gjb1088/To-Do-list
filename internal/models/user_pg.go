package models

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserStorePostgres struct {
	db *sqlx.DB
}

func NewUserStorePostgres(db *sqlx.DB) *UserStorePostgres {
	return &UserStorePostgres{db: db}
}

func (s *UserStorePostgres) Create(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO users(username,password_hash) VALUES($1,$2)`,
		username, hash,
	)
	return err
}

func (s *UserStorePostgres) Authenticate(username, password string) bool {
	var hash []byte
	if err := s.db.Get(&hash, `SELECT password_hash FROM users WHERE username=$1`, username); err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}
