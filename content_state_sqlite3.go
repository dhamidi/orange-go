package main

import (
	"database/sql"
	"fmt"
	"strings"
)

var _ ContentState = (*PersistentContentState)(nil)

// PersistentContentState is a **sketch** of a persistent version of storing all content.
//
// Using a persistent version means we do not need to replay the
// entire log on application startup, but instead can *remember* our
// position and resume from there.
//
// This particular implementation uses sqlite3 for this purpose.
//
// Since storing graph-based data in a SQL database is not straightforward,
// I eventually gave up, because at this stage it's not necessary.
//
// One day I'll return and implement the rest, just to demonstrate
// that you don't have to keep the entire application state in memory.
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
		"CREATE TABLE IF NOT EXISTS submissions (item_id TEXT PRIMARY KEY, submitter TEXT, url TEXT, title TEXT, submitted_at TIMESTAMP);",
		"CREATE TABLE IF NOT EXISTS submission_previews (item_id TEXT PRIMARY KEY, title TEXT, image_url TEXT, description TEXT, generated_at TIMESTAMP);",
		"CREATE TABLE IF NOT EXISTS submission_votes (item_id TEXT, voter TEXT, voted_at TIMESTAMP, PRIMARY KEY (item_id, voter));",
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

func (self *PersistentContentState) GetSubmission(itemID string) (*Submission, error) {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	submission := &Submission{
		ItemID: itemID,
	}
	if err := db.QueryRow(`SELECT submitter, url, title, submitted_at FROM submissions WHERE item_id = ?`, itemID).Scan(
		&submission.Submitter, &submission.Url, &submission.Title, &submission.SubmittedAt); err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}
	return submission, nil
}

func (self *PersistentContentState) PutComment(comment *Comment) error {
	return ErrItemNotFound
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

func (self *PersistentContentState) TopNSubmissions(n int, after int) ([]*Submission, error) {
	db, err := sql.Open("sqlite3", self.conninfo())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT item_id, submitter, url, title, submitted_at FROM submissions ORDER BY submitted_at DESC LIMIT ? OFFSET ?`, n, after)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions: %w", err)
	}
	defer rows.Close()
	submissions := make([]*Submission, 0, n)

	for rows.Next() {
		submission := &Submission{}
		if err := rows.Scan(&submission.ItemID, &submission.Submitter, &submission.Url, &submission.Title, &submission.SubmittedAt); err != nil {
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
    INSERT OR REPLACE INTO submission_votes (item_id, voter, voted_at) 
          values (?, ?, ?) 
    ON CONFLICT (item_id, voter)
    DO NOTHING
  `, vote.For, vote.By, vote.At)
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
  SELECT item_id FROM submission_votes WHERE voter = ? AND item_id IN (%s)
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

func (self *PersistentContentState) GetActiveSubscribers() ([]string, error) { return []string{}, nil }

func (self *PersistentContentState) GetSubmissionForComment(commentID TreeID) (*Submission, error) {
	panic("unimplemented")
}

func (self *PersistentContentState) GetSubscriptionSettings(username string) (*SubscriptionSettings, error) {
	panic("unimplemented")
}

func (self *PersistentContentState) PutSubscriptionSettings(settings *SubscriptionSettings) error {
	panic("unimplemented")
}
