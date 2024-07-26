package main

import (
	"errors"
	"time"
)

var (
	ErrUnknownSubmission = errors.New("unknown submission")
	ErrMissingVoter      = errors.New("missing voter")
	ErrAlreadyVoted      = errors.New("already voted")
)

type UpvoteSubmission struct {
	ItemID  string
	Voter   string
	VotedAt time.Time
}

func (cmd *UpvoteSubmission) CommandName() string { return "UpvoteSubmission" }

func init() {
	DefaultSerializer.Register("UpvoteSubmission", func() Command { return &UpvoteSubmission{} })
}

func (self *Content) handleUpvoteSubmission(cmd *UpvoteSubmission) error {
	if cmd.ItemID == "" {
		return ErrMissingItemID
	}
	if cmd.Voter == "" {
		return ErrMissingVoter
	}

	votes, err := self.state.HasVotedFor(cmd.Voter, []string{cmd.ItemID})
	if err != nil {
		return err
	}
	if len(votes) > 0 && votes[0] {
		return ErrAlreadyVoted
	}

	return self.state.RecordVote(&Vote{
		For: cmd.ItemID,
		By:  cmd.Voter,
		At:  cmd.VotedAt,
	})
}
