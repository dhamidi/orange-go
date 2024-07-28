package main

type FindSubmission struct {
	ItemID     string
	Submission *Submission
}

func (q *FindSubmission) QueryName() string { return "FindSubmission" }

func NewFindSubmission(itemID string) *FindSubmission {
	return &FindSubmission{ItemID: itemID}
}

func (self *Content) findSubmission(q *FindSubmission) error {
	submission, err := self.state.GetSubmission(q.ItemID)
	if err != nil {
		return err
	}
	q.Submission = submission
	return nil
}
