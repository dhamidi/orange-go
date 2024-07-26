package main

import (
	"database/sql"
	"fmt"
	"strings"
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
		"CREATE TABLE IF NOT EXISTS submission_votes (item_id TEXT, voter TEXT, voted_at TIMESTAMP);",
		"CREATE INDEX IF NOT EXISTS submission_voters ON submission_votes (item_id, voter);",
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

func (self *PersistentContentState) RecordVote(vote *Vote) error {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	_, err = db.Query(`
    INSERT INTO submission_votes (item_id, voter_id, voted_at) values (?, ?, ?) 
    WHERE (select count(*) from submission_votes WHERE item_id = ? and voter_id = ?) = 0
  `, vote.For, vote.By, vote.At, vote.For, vote.By)
	if err != nil {
		return fmt.Errorf("failed to insert vote: %w", err)
	}

	return nil
}

func (self *PersistentContentState) HasVotedFor(user string, itemIDs []string) ([]bool, error) {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	idList := []string{}
	for _, id := range itemIDs {
		idList = append(idList, fmt.Sprintf("%q", id))
	}
	ids := strings.Join(idList, ",")

	rows, err := db.Query(fmt.Sprintf(`
  SELECT item_id FROM submission_votes WHERE voter_id = ? AND item_id IN (%s)
  `, ids), user)
	if err != nil {
		return nil, fmt.Errorf("failed to get voting state: %w", err)
	}

	voted := make([]bool, len(itemIDs))
	votedFor := map[string]bool{}
	for rows.Next() {
		itemID := ""
		rows.Scan(&itemID)
		votedFor[itemID] = true
	}

	for i, itemID := range itemIDs {
		voted[i] = votedFor[itemID]
	}

	return voted, nil
}
