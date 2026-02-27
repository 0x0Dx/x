package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	path string
}

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

func (d *DB) Set(table, key, value string) error {
	_, err := d.Exec(fmt.Sprintf("INSERT OR REPLACE INTO %s (key, value) VALUES (?, ?)", table), key, value)
	return err
}

func (d *DB) Delete(table, key string) error {
	_, err := d.Exec(fmt.Sprintf("DELETE FROM %s WHERE key = ?", table), key)
	return err
}

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

func (d *DB) Close() error {
	return d.DB.Close()
}

func EnsureDB(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return New(path)
}
