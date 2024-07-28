package main

import (
	"errors"
	"time"
)

const MAX_COMMENT_LENGTH_IN_CHARACTERS = 300
const MIN_COMMENT_LENGTH_IN_CHARACTERS = 3

var (
	ErrUncommentableItem = errors.New("uncommentable item")
	ErrCommentTooLong    = errors.New("comment too long")
	ErrCommentTooShort   = errors.New("comment too short")
)

type PostComment struct {
	ParentID TreeID // comment or submission ID
	Author   string
	PostedAt time.Time
	Content  string
}

func (cmd *PostComment) CommandName() string { return "PostComment" }

func init() {
	DefaultSerializer.Register("PostComment", func() Command { return new(PostComment) })
}

func (self *Content) handlePostComment(cmd *PostComment) error {
	if len(cmd.Content) > MAX_COMMENT_LENGTH_IN_CHARACTERS {
		return ErrCommentTooLong
	}

	if len(cmd.Content) < MIN_COMMENT_LENGTH_IN_CHARACTERS {
		return ErrCommentTooShort
	}

	comment := &Comment{
		ParentID: cmd.ParentID,
		Author:   cmd.Author,
		Content:  cmd.Content,
		PostedAt: cmd.PostedAt,
	}

	if err := self.state.PutComment(comment); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return ErrUncommentableItem
		}
		return err
	}
	return nil
}
