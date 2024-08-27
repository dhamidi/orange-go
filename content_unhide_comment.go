package main

import (
	"fmt"
	"time"
)

type UnhideComment struct {
	CommentID  TreeID
	UnhiddenAt time.Time
	UnhiddenBy string
}

func (cmd *UnhideComment) CommandName() string { return "UnhideComment" }

func init() {
	DefaultCommandRegistry.Register("UnhideComment", func() Command { return new(UnhideComment) })
}

func (self *Content) handleUnhideComment(cmd *UnhideComment) error {
	submission, err := self.state.GetSubmissionForComment(cmd.CommentID)
	if err != nil {
		return fmt.Errorf("failed to get submission for comment %q: %w", cmd.CommentID, err)
	}
	comment := submission.Comment(cmd.CommentID)
	if comment == nil {
		return ErrItemNotFound
	}
	comment.Hidden = false
	return nil
}
