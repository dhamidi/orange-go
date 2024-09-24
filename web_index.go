package main

import (
	"net/http"
	"net/url"
	"orange/pages"
	"strconv"
)

func (web *WebApp) PageIndex(w http.ResponseWriter, req *http.Request) {
	pageData := web.PageData(req)
	q := NewFrontpageQuery(pageData.Username())
	if after, err := strconv.Atoi(req.URL.Query().Get("after")); err == nil {
		q.After = after
	}

	if err := web.app.HandleQuery(q); err != nil {
		http.Error(w, "failed to load front page", http.StatusInternalServerError)
		return
	}
	templateData := []*pages.Submission{}
	index := 1 + q.After
	for _, submission := range q.Submissions {
		title := ""
		if submission.Preview != nil {
			if submission.Preview.Title != nil {
				title = *submission.Preview.Title
			}
		}

		if submission.Hidden && !pageData.IsAdmin {
			continue
		}

		if len(templateData) >= 10 {
			break
		}
		templateData = append(templateData, &pages.Submission{
			Index:          uint64(index),
			ItemID:         submission.ItemID,
			Title:          submission.Title,
			GeneratedTitle: title,
			Hidden:         submission.Hidden,
			Url:            submission.Url,
			SubmittedAt:    submission.SubmittedAt,
			Submitter:      submission.Submitter,
			VoteCount:      submission.VoteCount,
			CommentCount:   submission.CommentCount,
			CanVote:        !submission.ViewerHasVoted,
		})
		index++
	}

	if len(templateData) >= 10 {
		pageData.LoadMore = &url.URL{Path: req.URL.Path}
		pageData.LoadMore.RawQuery = (&url.Values{"after": []string{strconv.Itoa(q.After + 10)}}).Encode()
	}

	if isHX(req) {
		pages.SubmissionList(templateData, pageData.LoadMore, pageData.IsAdmin).Render(w)
		return
	}

	_ = pages.IndexPage(req.URL.Path, templateData, pageData).Render(w)
}
