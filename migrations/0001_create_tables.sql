-- migrations/0001_create_tables.sql

-- users table
CREATE TABLE IF NOT EXISTS users (
  username      TEXT PRIMARY KEY,
  password_hash BYTEA NOT NULL
);

-- todos table
CREATE TABLE IF NOT EXISTS todos (
  id        SERIAL     PRIMARY KEY,
  title     TEXT       NOT NULL,
  completed BOOLEAN    NOT NULL DEFAULT FALSE,
  username  TEXT       NOT NULL REFERENCES users(username)
);
