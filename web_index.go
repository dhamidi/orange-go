package main

import (
	"net/http"
	"orange/pages"
)

func (web *WebApp) PageIndex(w http.ResponseWriter, req *http.Request) {
	pageData := web.PageData(req)

	q := NewFrontpageQuery(pageData.Username())
	if err := web.app.HandleQuery(q); err != nil {
		http.Error(w, "failed to load front page", http.StatusInternalServerError)
		return
	}
	templateData := []*pages.Submission{}
	for _, submission := range q.Submissions {
		title := ""
		imageURL := (*string)(nil)
		if submission.Preview != nil {
			if submission.Preview.Title != nil {
				title = *submission.Preview.Title
			}
			imageURL = submission.Preview.ImageURL
		}
		templateData = append(templateData, &pages.Submission{
			ItemID:         submission.ItemID,
			Title:          submission.Title,
			GeneratedTitle: title,
			ImageURL:       imageURL,
			Url:            submission.Url,
			SubmittedAt:    submission.SubmittedAt,
			Submitter:      submission.Submitter,
			VoteCount:      submission.VoteCount,
			CommentCount:   submission.CommentCount,
			CanVote:        !submission.ViewerHasVoted,
		})
	}

	if isHX(req) {
		pages.SubmissionList(templateData).Render(w)
		return
	}

	_ = pages.IndexPage(req.URL.Path, templateData, web.PageData(req)).Render(w)
}
