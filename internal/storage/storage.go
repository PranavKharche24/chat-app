package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

// NewDatabase opens/creates the SQLite DB and users table.
func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	schema := `
	CREATE TABLE IF NOT EXISTS users (
	  id         INTEGER PRIMARY KEY AUTOINCREMENT,
	  username   TEXT    UNIQUE NOT NULL,
	  email      TEXT    UNIQUE NOT NULL,
	  dob        TEXT    NOT NULL,
	  full_name  TEXT    NOT NULL,
	  password   TEXT    NOT NULL
	);`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return &Database{db}, nil
}

// RegisterUser creates a new user and returns a zero-padded 4-digit ID.
func (d *Database) RegisterUser(username, email, dob, fullname, pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	res, err := d.db.Exec(
		`INSERT INTO users(username,email,dob,full_name,password) VALUES(?,?,?,?,?)`,
		username, email, dob, fullname, string(hash),
	)
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	return fmt.Sprintf("%04d", id), nil
}

// Authenticate verifies a numeric userID and password.
func (d *Database) Authenticate(userID, pass string) error {
	var stored string
	if err := d.db.QueryRow(`SELECT password FROM users WHERE id = ?`, userID).Scan(&stored); err != nil {
		return errors.New("user not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(stored), []byte(pass)) != nil {
		return errors.New("invalid password")
	}
	return nil
}

// LookupByUsername finds the userID for a given username.
func (d *Database) LookupByUsername(username string) (string, error) {
	var id int
	if err := d.db.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id); err != nil {
		return "", errors.New("user not found")
	}
	return fmt.Sprintf("%04d", id), nil
}

// GetProfile returns username,email,dob,full_name for SETTINGS.
func (d *Database) GetProfile(userID string) (u, e, dob, fn string, err error) {
	err = d.db.QueryRow(
		`SELECT username,email,dob,full_name FROM users WHERE id = ?`, userID,
	).Scan(&u, &e, &dob, &fn)
	return
}
