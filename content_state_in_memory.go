package main

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"time"
)

type InMemoryContentState struct {
	FrontpageDirty   bool
	LastSubmissionAt time.Time
	Submissions      []*Submission
	VotesByItemID    map[string][]string
}

func (self *InMemoryContentState) scoreSubmissions() {
	for _, submission := range self.Submissions {
		self.score(submission)
	}
}

func (self *InMemoryContentState) score(s *Submission) {
	voteCount := len(self.VotesByItemID[s.ItemID])
	age := float64(self.LastSubmissionAt.Sub(s.SubmittedAt)) / float64(24*time.Hour)
	decay := float32(math.Pow(0.9, age))
	s.Score = float32(voteCount) * decay
}

func (self *InMemoryContentState) refreshFrontpage() {
	if !self.FrontpageDirty {
		return
	}
	self.scoreSubmissions()
	slices.SortFunc(self.Submissions, func(i, j *Submission) int {
		a, b := i.Score, j.Score
		if a == b {
			return j.SubmittedAt.Compare(i.SubmittedAt)
		}
		return cmp.Compare(b, a)
	})
	self.FrontpageDirty = false
}

func (self *InMemoryContentState) PutSubmissionPreview(preview *SubmissionPreview) error {
	for _, submission := range self.Submissions {
		if submission.ItemID == preview.ItemID {
			submission.Preview = preview
			return nil
		}
	}
	return nil
}

func NewInMemoryContentState() *InMemoryContentState {
	return &InMemoryContentState{
		Submissions:   make([]*Submission, 0),
		VotesByItemID: map[string][]string{},
	}
}

func (self *InMemoryContentState) GetSubmission(itemID string) (*Submission, error) {
	for _, submission := range self.Submissions {
		if submission.ItemID == itemID {
			return submission, nil
		}
	}
	return nil, ErrItemNotFound
}

func (self *InMemoryContentState) PutSubmission(submission *Submission) error {
	self.Submissions = append(self.Submissions, submission)
	if submission.SubmittedAt.After(self.LastSubmissionAt) {
		self.LastSubmissionAt = submission.SubmittedAt
	}
	self.FrontpageDirty = true
	return nil
}

func (self *InMemoryContentState) PutComment(comment *Comment) error {
	submissionID := comment.ParentID[0]
	submission := (*Submission)(nil)
	for _, s := range self.Submissions {
		if s.ItemID == submissionID {
			submission = s
			break
		}
	}

	if submission == nil {
		return ErrItemNotFound
	}

	if len(comment.ParentID) == 1 {
		comment.Index = len(submission.Comments)
		submission.Comments = append(submission.Comments, comment)
		submission.CommentCount++
		return nil
	}

	currentComments := ([]*Comment)(submission.Comments)
	parentComment := (*Comment)(nil)
	for _, index := range comment.ParentID[1:] {
		i := 0
		if _, err := fmt.Sscanf(index, "%d", &i); err != nil {
			return fmt.Errorf("Malformed tree ID part %q: %w", index, ErrMalformedTreeID)
		}

		parentComment = currentComments[i]
		if comment == nil {
			return fmt.Errorf("No comment at index %d: %w", i, ErrItemNotFound)
		}
		currentComments = parentComment.Children
	}

	parentComment.AddChild(comment)
	submission.CommentCount++
	return nil
}

func (self *InMemoryContentState) TopNSubmissions(n int) ([]*Submission, error) {
	if n > len(self.Submissions) {
		n = len(self.Submissions)
	}
	self.refreshFrontpage()
	topN := self.Submissions[:n]
	for _, s := range topN {
		s.VoteCount = len(self.VotesByItemID[s.ItemID])
	}
	return topN, nil
}

func (self *InMemoryContentState) RecordVote(vote *Vote) error {
	voters, ok := self.VotesByItemID[vote.For]
	if !ok {
		voters = []string{}
	}

	if slices.Contains(voters, vote.By) {
		return nil
	}

	voters = append(voters, vote.By)
	self.VotesByItemID[vote.For] = voters
	self.FrontpageDirty = true
	return nil
}

func (self *InMemoryContentState) HasVotedFor(user string, itemIDs []string) ([]bool, error) {
	result := make([]bool, len(itemIDs))
	for i, itemID := range itemIDs {
		voters, ok := self.VotesByItemID[itemID]
		if !ok {
			continue
		}
		result[i] = slices.Contains(voters, user)
	}
	return result, nil
}
