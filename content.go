package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PostLink struct {
	Submitter   string
	Url         string
	Title       string
	SubmittedAt time.Time
}

func (cmd *PostLink) CommandName() string {
	return "PostLink"
}

func init() {
	DefaultSerializer.Register("PostLink", func() Command { return &PostLink{} })
}

type ContentState interface {
	PutSubmission(submission *Submission) error
	TopNSubmissions(n int) ([]*Submission, error)
}

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
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS submissions (submitter TEXT, url TEXT, title TEXT, submitted_at TIMESTAMP)"); err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
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
	if _, err := db.Exec(`INSERT OR REPLACE INTO submissions (submitter, url, title, submitted_at) VALUES (?, ?, ?, ?)`,
		submission.Submitter, submission.Url, submission.Title, submission.SubmittedAt); err != nil {
		return fmt.Errorf("failed to insert submission: %w", err)
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
	submissions := make([]*Submission, 0)
	for rows.Next() {
		submission := &Submission{}
		if err := rows.Scan(&submission.Submitter, &submission.Url, &submission.Title, &submission.SubmittedAt); err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}
		submissions = append(submissions, submission)
	}
	return submissions, nil
}

type Submission struct {
	Submitter   string
	Url         string
	Title       string
	SubmittedAt time.Time
}

type InMemoryContentState struct {
	Submissions []*Submission
}

func NewInMemoryContentState() *InMemoryContentState {
	return &InMemoryContentState{
		Submissions: make([]*Submission, 0),
	}
}

func (self *InMemoryContentState) PutSubmission(submission *Submission) error {
	self.Submissions = append(self.Submissions, submission)
	return nil
}

func (self *InMemoryContentState) TopNSubmissions(n int) ([]*Submission, error) {
	if n > len(self.Submissions) {
		n = len(self.Submissions)
	}
	return self.Submissions[:n], nil
}

type Content struct {
	state ContentState
}

func NewContent(state ContentState) *Content {
	return &Content{state: state}
}

func NewDefaultContent() *Content {
	return NewContent(NewInMemoryContentState())
}

func (self *Content) HandleCommand(cmd Command) error {
	switch cmd := cmd.(type) {
	case *PostLink:
		return self.handlePostLink(cmd)
	}
	return ErrCommandNotAccepted
}

var (
	EmptyTitle = errors.New("title cannot be empty")
	EmptyUrl   = errors.New("url cannot be empty")
)

func (self *Content) handlePostLink(cmd *PostLink) error {
	if cmd.Title == "" {
		return EmptyTitle
	}

	if cmd.Url == "" {
		return EmptyUrl
	}

	return self.state.PutSubmission(&Submission{
		Submitter:   cmd.Submitter,
		Url:         cmd.Url,
		Title:       cmd.Title,
		SubmittedAt: cmd.SubmittedAt,
	})
}

type GetFrontpageSubmissions struct {
	Submissions []*Submission
}

func (q *GetFrontpageSubmissions) QueryName() string {
	return "GetFrontpageSubmissions"
}

func (self *Content) HandleQuery(query Query) error {
	switch query := query.(type) {
	case *GetFrontpageSubmissions:
		submissions, err := self.state.TopNSubmissions(10)
		if err != nil {
			return err
		}
		query.Submissions = submissions
	default:
		return ErrQueryNotAccepted
	}
	return nil
}

func NewFrontpageQuery() *GetFrontpageSubmissions {
	return &GetFrontpageSubmissions{
		Submissions: []*Submission{},
	}
}
