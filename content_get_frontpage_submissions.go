package main

type GetFrontpageSubmissions struct {
	Submissions []*Submission
}

func (q *GetFrontpageSubmissions) QueryName() string {
	return "GetFrontpageSubmissions"
}

func NewFrontpageQuery() *GetFrontpageSubmissions {
	return &GetFrontpageSubmissions{
		Submissions: []*Submission{},
	}
}

func (self *Content) getFrontpageSubmissions(query *GetFrontpageSubmissions) error {
	submissions, err := self.state.TopNSubmissions(10)
	if err != nil {
		return err
	}
	query.Submissions = submissions
	return nil
}
