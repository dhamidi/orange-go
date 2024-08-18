package main

type GetFrontpageSubmissions struct {
	Viewer *string
	After  int

	Submissions []*Submission
}

func (q *GetFrontpageSubmissions) QueryName() string {
	return "GetFrontpageSubmissions"
}

func NewFrontpageQuery(viewer *string) *GetFrontpageSubmissions {
	return &GetFrontpageSubmissions{
		Viewer:      viewer,
		Submissions: []*Submission{},
	}
}

func (self *Content) getFrontpageSubmissions(query *GetFrontpageSubmissions) error {
	submissions, err := self.state.TopNSubmissions(10, query.After)
	if err != nil {
		return err
	}
	if query.Viewer != nil {
		viewer := *query.Viewer
		itemIDs := make([]string, len(submissions))
		for i, s := range submissions {
			itemIDs[i] = s.ItemID
		}
		voted, _ := self.state.HasVotedFor(viewer, itemIDs)
		for i, s := range submissions {
			s.ViewerHasVoted = voted[i]
		}
	}
	query.Submissions = submissions
	return nil
}
