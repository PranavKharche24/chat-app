package storage

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

// Database wraps an SQLite connection.
type Database struct {
	db *sql.DB
}

// NewDatabase opens (or creates) an SQLite database at the given filepath and creates tables if needed.
func NewDatabase(filepath string) (*Database, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	// Create users table
	createUsers := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`
	// Create messages table
	createMessages := `CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		content TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createUsers); err != nil {
		return nil, err
	}
	if _, err := db.Exec(createMessages); err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

// RegisterUser hashes the password and stores a new user.
func (d *Database) RegisterUser(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("INSERT INTO users(username, password) VALUES(?,?)", username, string(hash))
	return err
}

// AuthenticateUser checks if the provided password matches the stored hash.
func (d *Database) AuthenticateUser(username, password string) error {
	var hash string
	row := d.db.QueryRow("SELECT password FROM users WHERE username = ?", username)
	if err := row.Scan(&hash); err != nil {
		return errors.New("user not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return errors.New("invalid password")
	}
	return nil
}

// Message represents a stored chat message.
type Message struct {
	Username  string
	Content   string
	Timestamp string
}

// StoreMessage saves a new message to the database.
func (d *Database) StoreMessage(username, content string) error {
	_, err := d.db.Exec("INSERT INTO messages(username, content) VALUES(?,?)", username, content)
	return err
}

// GetMessages retrieves the latest 'limit' messages.
func (d *Database) GetMessages(limit int) ([]Message, error) {
	rows, err := d.db.Query("SELECT username, content, timestamp FROM messages ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Username, &msg.Content, &msg.Timestamp); err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
