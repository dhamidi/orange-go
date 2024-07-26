package main

import (
	"database/sql"
	"fmt"
	"iter"

	_ "github.com/mattn/go-sqlite3"
)

type FileCommandLog struct {
	filename   string
	serializer Serializer
}

func NewFileCommandLog(filename string, serializer Serializer) *FileCommandLog {
	return &FileCommandLog{
		filename:   filename,
		serializer: serializer,
	}
}

func (f *FileCommandLog) conninfo() string {
	return fmt.Sprintf("file:%s?_journal=wal&_txlock=immediate", f.filename)
}

func (f *FileCommandLog) Setup() error {
	db, err := sql.Open("sqlite3", f.conninfo())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS commands (id INTEGER PRIMARY KEY, message TEXT)"); err != nil {
		return fmt.Errorf("failed to create commands table: %w", err)
	}

	return nil
}

func (f *FileCommandLog) Append(command Command) error {
	db, err := sql.Open("sqlite3", f.conninfo())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	encoded, err := f.serializer.Encode(command)
	if err != nil {
		return fmt.Errorf("failed to encode command: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec("INSERT INTO commands (message) VALUES (?)", encoded); err != nil {
		return fmt.Errorf("failed to insert command: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (f *FileCommandLog) After(ID int) (iter.Seq[*PersistedCommand], error) {
	db, err := sql.Open("sqlite3", f.conninfo())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return func(yield func(*PersistedCommand) bool) {
		defer db.Close()

		query := `SELECT id, message FROM commands WHERE id > ? ORDER BY id`
		rows, err := db.Query(query, ID)
		if err != nil {
			panic(fmt.Errorf("failed to query commands: %w", err))
		}

		for rows.Next() {
			var (
				id      int
				message []byte
			)

			if err := rows.Scan(&id, &message); err != nil {
				panic(fmt.Errorf("failed to scan row: %w", err))
			}

			cmd := new(Command)
			if err := f.serializer.Decode(message, cmd); err != nil {
				panic(fmt.Errorf("failed to decode command: %w (raw: %s)", err, message))
			}
			if shouldContinue := yield(&PersistedCommand{ID: id, Message: *cmd}); shouldContinue == false {
				break
			}
		}
	}, nil
}
