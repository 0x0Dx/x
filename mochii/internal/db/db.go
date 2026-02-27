package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database connection.
type DB struct {
	*sql.DB
	path string
}

// New opens or creates a SQLite database at the given path.
func New(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	d := &DB{db, path}
	if err := d.init(); err != nil {
		return nil, err
	}

	return d, nil
}

// init creates the required database tables.
func (d *DB) init() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS refs (key TEXT PRIMARY KEY, value TEXT)`,
		`CREATE TABLE IF NOT EXISTS pkginst (key TEXT PRIMARY KEY, value TEXT)`,
		`CREATE TABLE IF NOT EXISTS prebuilts (key TEXT PRIMARY KEY, value TEXT)`,
		`CREATE TABLE IF NOT EXISTS netsources (key TEXT PRIMARY KEY, value TEXT)`,
	}

	for _, t := range tables {
		if _, err := d.Exec(t); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	return nil
}

// Get retrieves a value from a table by key. Returns (value, found, error).
func (d *DB) Get(table, key string) (string, bool, error) {
	var value string
	err := d.QueryRow(fmt.Sprintf("SELECT value FROM %s WHERE key = ?", table), key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

// Set inserts or updates a key-value pair in a table.
func (d *DB) Set(table, key, value string) error {
	_, err := d.Exec(fmt.Sprintf("INSERT OR REPLACE INTO %s (key, value) VALUES (?, ?)", table), key, value)
	return err
}

// Delete removes a key from a table.
func (d *DB) Delete(table, key string) error {
	_, err := d.Exec(fmt.Sprintf("DELETE FROM %s WHERE key = ?", table), key)
	return err
}

// List returns all key-value pairs from a table.
func (d *DB) List(table string) (map[string]string, error) {
	rows, err := d.Query(fmt.Sprintf("SELECT key, value FROM %s", table))
	if err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.DB.Close()
}

// EnsureDB creates the database directory if needed, then opens the database.
func EnsureDB(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return New(path)
}
