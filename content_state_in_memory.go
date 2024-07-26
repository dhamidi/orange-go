package main

import "slices"

type InMemoryContentState struct {
	Submissions   []*Submission
	VotesByItemID map[string][]string
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

func (self *InMemoryContentState) PutSubmission(submission *Submission) error {
	self.Submissions = append(self.Submissions, submission)
	slices.SortFunc(self.Submissions, func(i, j *Submission) int {
		if i.SubmittedAt.After(j.SubmittedAt) {
			return -1
		} else if i.SubmittedAt.Before(j.SubmittedAt) {
			return 1
		} else {
			return 0
		}
	})
	return nil
}

func (self *InMemoryContentState) TopNSubmissions(n int) ([]*Submission, error) {
	if n > len(self.Submissions) {
		n = len(self.Submissions)
	}
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
