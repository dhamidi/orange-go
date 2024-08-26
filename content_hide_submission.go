package main

import (
	"errors"
	"fmt"
	"time"
)

type HideSubmission struct {
	ItemID   string
	HiddenAt time.Time
	HiddenBy string
}

func (cmd *HideSubmission) CommandName() string { return "HideSubmission" }

func init() {
	DefaultCommandRegistry.Register("HideSubmission", func() Command { return new(HideSubmission) })
}

func (self *Content) handleHideSubmission(cmd *HideSubmission) error {
	submission, err := self.state.GetSubmission(cmd.ItemID)
	if errors.Is(err, ErrItemNotFound) {
		return err
	}
	if err != nil {
		return fmt.Errorf("failed to get submission %q: %w", cmd.ItemID, err)
	}
	submission.Hidden = true
	return self.state.PutSubmission(submission)
}
