package main

import (
	"errors"
	"time"
)

type ContentState interface {
	PutSubmissionPreview(preview *SubmissionPreview) error
	PutSubmission(submission *Submission) error
	TopNSubmissions(n int) ([]*Submission, error)
}

type Submission struct {
	ItemID      string
	Submitter   string
	Url         string
	Title       string
	SubmittedAt time.Time
	Preview     *SubmissionPreview
}

type SubmissionPreview struct {
	ItemID      string
	GeneratedAt time.Time
	Title       *string
	Description *string
	ImageURL    *string
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
	case *SetSubmissionPreview:
		return self.handleSetSubmissionPreview(cmd)
	}
	return ErrCommandNotAccepted
}

var (
	ErrEmptyTitle    = errors.New("title cannot be empty")
	ErrEmptyUrl      = errors.New("url cannot be empty")
	ErrMalformedURL  = errors.New("url is malformed")
	ErrMissingItemID = errors.New("item ID is missing")
)

func (self *Content) HandleQuery(query Query) error {
	switch query := query.(type) {
	case *GetFrontpageSubmissions:
		return self.getFrontpageSubmissions(query)
	default:
		return ErrQueryNotAccepted
	}
}
