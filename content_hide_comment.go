package main

import (
	"fmt"
	"time"
)

type HideComment struct {
	CommentID TreeID
	HiddenAt  time.Time
	HiddenBy  string
}

func (cmd *HideComment) CommandName() string { return "HideComment" }

func init() {
	DefaultCommandRegistry.Register("HideComment", func() Command { return new(HideComment) })
}

func (self *Content) handleHideComment(cmd *HideComment) error {
	submission, err := self.state.GetSubmissionForComment(cmd.CommentID)
	if err != nil {
		return fmt.Errorf("failed to get submission for comment %q: %w", cmd.CommentID, err)
	}
	comment := submission.Comment(cmd.CommentID)
	if comment == nil {
		return ErrItemNotFound
	}
	comment.Hidden = true
	return nil
}
