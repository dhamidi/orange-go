package main

import (
	"errors"
	"strconv"
	"time"
)

type ContentState interface {
	PutSubmissionPreview(preview *SubmissionPreview) error
	PutSubmission(submission *Submission) error
	GetSubmission(itemID string) (*Submission, error)
	TopNSubmissions(n int, after int) ([]*Submission, error)
	RecordVote(vote *Vote) error
	HasVotedFor(user string, itemIDs []string) ([]bool, error)
	PutComment(comment *Comment) error
	GetSubmissionForComment(commentID TreeID) (*Submission, error)

	GetActiveSubscribers() ([]string, error)
	GetSubscriptionSettings(username string) (*SubscriptionSettings, error)
	PutSubscriptionSettings(settings *SubscriptionSettings) error
}

type Submission struct {
	ItemID         string
	Submitter      string
	Url            string
	Title          string
	SubmittedAt    time.Time
	Preview        *SubmissionPreview
	Hidden         bool
	VoteCount      int
	Score          float32
	ViewerHasVoted bool
	CommentCount   int
	Comments       []*Comment
}

func (s *Submission) Comment(id TreeID) *Comment {
	moves := id[1:]
	current := s.Comments
	for ci, move := range moves {
		i, _ := strconv.Atoi(move)
		if i >= len(current) {
			return nil
		}
		if ci == len(moves)-1 {
			return current[i]
		}

		current = current[i].Children
	}

	return nil
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
	ParentID    TreeID
	Content     string
	ContentHTML string
	Author      string
	PostedAt    time.Time
	Hidden      bool
	Index       int
	Children    []*Comment
}

func (c *Comment) IsHidden() bool          { return c.Hidden }
func (c *Comment) CommentID() string       { return c.ID().String() }
func (c *Comment) CommentParentID() string { return c.ParentID.String() }
func (c *Comment) CommentableID() string   { return c.ID().String() }
func (c *Comment) WrittenAt() time.Time    { return c.PostedAt }
func (c *Comment) CommentAuthor() string   { return c.Author }
func (c *Comment) CommentContent() string {
	if c.Hidden {
		return "[hidden]"
	}
	if c.ContentHTML != "" {
		return c.ContentHTML
	}

	c.ContentHTML = ConvertContentToHTML(c.Content)
	if c.ContentHTML == "" {
		return c.Content
	}

	return c.ContentHTML
}
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
	case *HideSubmission:
		return self.handleHideSubmission(cmd)
	case *UnhideSubmission:
		return self.handleUnhideSubmission(cmd)
	case *HideComment:
		return self.handleHideComment(cmd)
	case *UnhideComment:
		return self.handleUnhideComment(cmd)
	case *EnableSubscriptions:
		return self.handleEnableSubscriptions(cmd)
	case *DisableSubscriptions:
		return self.handleDisableSubscriptions(cmd)
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
	case *SubscriptionSettingsForUser:
		return self.findSubscriptionSettings(query)
	case *FindSubscribersForNewSubmission:
		return self.findSubscribersForNewSubmission(query)
	case *FindSubscribersForNewComment:
		return self.findSubscribersForNewComment(query)
	default:
		return ErrQueryNotAccepted
	}
}

func (self *Content) Inspect() map[string]any {
	return map[string]any{
		"content": self.state,
	}
}
