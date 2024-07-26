package main

import (
	"database/sql"
	"fmt"
)

type PersistentContentState struct {
	filename string
}

func NewPersistentContentState(filename string) *PersistentContentState {
	return &PersistentContentState{filename: filename}
}

func (self *PersistentContentState) Setup() error {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	schema := []string{
		"CREATE TABLE IF NOT EXISTS submissions (item_id TEXT, submitter TEXT, url TEXT, title TEXT, submitted_at TIMESTAMP);",
		"CREATE TABLE IF NOT EXISTS submission_previews (item_id TEXT, title TEXT, image_url TEXT, description TEXT, generated_at TIMESTAMP);",
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Commit()
	for _, stmt := range schema {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to create database schema: %w", err)
		}
	}
	return nil
}

func (self *PersistentContentState) conninfo() string {
	return fmt.Sprintf("file:%s?_journal=wal", self.filename)
}

func (self *PersistentContentState) PutSubmission(submission *Submission) error {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT OR REPLACE INTO submissions (item_id, submitter, url, title, submitted_at) VALUES (?, ?, ?, ?, ?)`,
		submission.ItemID, submission.Submitter, submission.Url, submission.Title, submission.SubmittedAt); err != nil {
		return fmt.Errorf("failed to insert submission: %w", err)
	}
	return nil
}

func (self *PersistentContentState) PutSubmissionPreview(preview *SubmissionPreview) error {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT OR REPLACE INTO submission_previews (item_id, title, image_url, description, generated_at) VALUES (?, ?, ?, ?, ?)`,
		preview.ItemID, preview.Title, preview.ImageURL, preview.Description, preview.GeneratedAt); err != nil {
		return fmt.Errorf("failed to insert submission preview: %w", err)
	}
	return nil
}

func (self *PersistentContentState) TopNSubmissions(n int) ([]*Submission, error) {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT submitter, url, title, submitted_at FROM submissions ORDER BY submitted_at DESC LIMIT ?`, n)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions: %w", err)
	}
	defer rows.Close()
	submissions := make([]*Submission, n)
	for rows.Next() {
		submission := &Submission{}
		if err := rows.Scan(&submission.Submitter, &submission.Url, &submission.Title, &submission.SubmittedAt); err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}
		preview := &SubmissionPreview{
			ItemID: submission.ItemID,
		}
		if err := db.QueryRow(
			"select title, image_url, description, generated_at from submission_previews where item_id = ?",
			submission.ItemID,
		).Scan(&preview.Title, &preview.ImageURL, &preview.Description, &preview.GeneratedAt); err == nil {
			submission.Preview = preview
		}

		submissions = append(submissions, submission)
	}
	return submissions, nil
}
