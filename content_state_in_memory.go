package main

import "slices"

type InMemoryContentState struct {
	Submissions []*Submission
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
		Submissions: make([]*Submission, 0),
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
	return self.Submissions[:n], nil
}
