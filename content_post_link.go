package main

import (
	"net/url"
	"time"
)

type PostLink struct {
	ItemID      string
	Submitter   string
	Url         string
	Title       string
	SubmittedAt time.Time
}

func (cmd *PostLink) CommandName() string {
	return "PostLink"
}

func init() {
	DefaultCommandRegistry.Register("PostLink", func() Command { return &PostLink{} })
}

func (self *Content) handlePostLink(cmd *PostLink) error {
	if cmd.Title == "" {
		return ErrEmptyTitle
	}

	if cmd.Url == "" {
		return ErrEmptyUrl
	}

	u, err := url.Parse(cmd.Url)
	if err != nil {
		return ErrMalformedURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrMalformedURL
	}

	if cmd.ItemID == "" {
		return ErrMissingItemID
	}

	return self.state.PutSubmission(&Submission{
		ItemID:      cmd.ItemID,
		Submitter:   cmd.Submitter,
		Url:         cmd.Url,
		Title:       cmd.Title,
		SubmittedAt: cmd.SubmittedAt,
	})
}
