package storage

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct{ db *sql.DB }

// NewDatabase opens or creates the SQLite file and the users table.
func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	createUsers := `
	CREATE TABLE IF NOT EXISTS users (
	  id         INTEGER PRIMARY KEY AUTOINCREMENT,
	  username   TEXT    UNIQUE NOT NULL,
	  email      TEXT    UNIQUE NOT NULL,
	  dob        TEXT    NOT NULL,
	  full_name  TEXT    NOT NULL,
	  hash       TEXT    NOT NULL
	)`
	if _, err := db.Exec(createUsers); err != nil {
		return nil, err
	}
	return &Database{db}, nil
}

// RegisterUser inserts a new user and returns their new userID.
func (d *Database) RegisterUser(username, email, dob, fullName, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := d.db.Exec(
		`INSERT INTO users(username,email,dob,full_name,hash) VALUES(?,?,?,?,?)`,
		username, email, dob, fullName, string(hash),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Authenticate checks a userID/password combo.
func (d *Database) Authenticate(userID int64, password string) error {
	var hash string
	err := d.db.QueryRow(
		`SELECT hash FROM users WHERE id = ?`, userID,
	).Scan(&hash)
	if err != nil {
		return errors.New("user not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return errors.New("invalid password")
	}
	return nil
}

// LookupByUsername returns the userID for a given username.
func (d *Database) LookupByUsername(username string) (int64, error) {
	var id int64
	err := d.db.QueryRow(
		`SELECT id FROM users WHERE username = ?`, username,
	).Scan(&id)
	if err != nil {
		return 0, errors.New("user not found")
	}
	return id, nil
}

// LookupUsernameByID returns the username for a given userID.
func (d *Database) LookupUsernameByID(userID int64) (string, error) {
	var username string
	err := d.db.QueryRow(
		`SELECT username FROM users WHERE id = ?`, userID,
	).Scan(&username)
	if err != nil {
		return "", errors.New("user not found")
	}
	return username, nil
}
