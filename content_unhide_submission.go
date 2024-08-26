package main

import (
	"errors"
	"fmt"
	"time"
)

type UnhideSubmission struct {
	ItemID     string
	UnhiddenAt time.Time
	UnhiddenBy string
}

func (cmd *UnhideSubmission) CommandName() string { return "UnhideSubmission" }

func init() {
	DefaultCommandRegistry.Register("UnhideSubmission", func() Command { return new(UnhideSubmission) })
}

func (self *Content) handleUnhideSubmission(cmd *UnhideSubmission) error {
	submission, err := self.state.GetSubmission(cmd.ItemID)
	if errors.Is(err, ErrItemNotFound) {
		return err
	}
	if err != nil {
		return fmt.Errorf("failed to get submission %q: %w", cmd.ItemID, err)
	}
	submission.Hidden = false
	return self.state.PutSubmission(submission)
}
