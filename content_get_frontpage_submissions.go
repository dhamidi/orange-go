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
	after := query.After
	result := []*Submission{}

	nonHiddenCount := 0
	for nonHiddenCount < 10 {
		moreSubmissions, err := self.state.TopNSubmissions(10, after)
		if err != nil {
			break
		}
		if len(moreSubmissions) == 0 {
			break
		}
		after += 10
		for _, s := range moreSubmissions {
			if anyOf(result, func(r *Submission) bool { return r.ItemID == s.ItemID }) {
				continue
			}
			if !s.Hidden {
				nonHiddenCount++
			}
			result = append(result, s)
		}
	}

	if query.Viewer != nil {
		viewer := *query.Viewer
		itemIDs := make([]string, len(result))
		for i, s := range result {
			itemIDs[i] = s.ItemID
		}
		voted, _ := self.state.HasVotedFor(viewer, itemIDs)
		for i, s := range result {
			s.ViewerHasVoted = voted[i]
		}
	}
	query.Submissions = result
	return nil
}
