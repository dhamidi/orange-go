package main

import (
	"errors"
	"time"
)

type ContentState interface {
	PutSubmissionPreview(preview *SubmissionPreview) error
	PutSubmission(submission *Submission) error
	GetSubmission(itemID string) (*Submission, error)
	TopNSubmissions(n int) ([]*Submission, error)
	RecordVote(vote *Vote) error
	HasVotedFor(user string, itemIDs []string) ([]bool, error)
	PutComment(comment *Comment) error
}

type Submission struct {
	ItemID         string
	Submitter      string
	Url            string
	Title          string
	SubmittedAt    time.Time
	Preview        *SubmissionPreview
	VoteCount      int
	Score          float32
	ViewerHasVoted bool
	CommentCount   int
	Comments       []*Comment
}

type SubmissionPreview struct {
	ItemID      string
	GeneratedAt time.Time
	Title       *string
	Description *string
	ImageURL    *string
}

type Vote struct {
	By  string
	For string
	At  time.Time
}

type Comment struct {
	ParentID TreeID
	Content  string
	Author   string
	PostedAt time.Time
	Index    int
	Children []*Comment
}

func (c *Comment) CommentableID() string  { return c.ID().String() }
func (c *Comment) WrittenAt() time.Time   { return c.PostedAt }
func (c *Comment) CommentAuthor() string  { return c.Author }
func (c *Comment) CommentContent() string { return c.Content }
func (c *Comment) AllChildren() []interface{} {
	asInterface := make([]interface{}, len(c.Children))
	for i := range c.Children {
		asInterface[i] = c.Children[i]
	}
	return asInterface
}

func (c *Comment) ID() TreeID {
	return c.ParentID.And(c.Index)
}

func (child *Comment) Of(parent *Comment) bool {
	parentID := parent.ID()

	for i := range child.ParentID {
		if child.ParentID[i] != parentID[i] {
			return false
		}
	}
	return true
}

func (parent *Comment) AddChild(child *Comment) {
	if !child.Of(parent) {
		return
	}

	childCount := len(parent.Children)
	child.Index = childCount
	parent.Children = append(parent.Children, child)
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
	case *UpvoteSubmission:
		return self.handleUpvoteSubmission(cmd)
	case *PostComment:
		return self.handlePostComment(cmd)
	}
	return ErrCommandNotAccepted
}

var (
	ErrEmptyTitle    = errors.New("title cannot be empty")
	ErrEmptyUrl      = errors.New("url cannot be empty")
	ErrMalformedURL  = errors.New("url is malformed")
	ErrMissingItemID = errors.New("item ID is missing")
	ErrItemNotFound  = errors.New("item not found")
)

func (self *Content) HandleQuery(query Query) error {
	switch query := query.(type) {
	case *GetFrontpageSubmissions:
		return self.getFrontpageSubmissions(query)
	case *FindSubmission:
		return self.findSubmission(query)
	default:
		return ErrQueryNotAccepted
	}
}
