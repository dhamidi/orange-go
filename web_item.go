package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) PageItem(w http.ResponseWriter, req *http.Request) {
	pageData := web.PageData(req)
	treeID := NewTreeID(req.FormValue("id"))
	q := NewFindSubmission(treeID.Root())
	if err := web.app.HandleQuery(q); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load submission", http.StatusInternalServerError)
		return
	}
	comments := []pages.Comment{}
	if len(treeID) > 1 {
		comment := q.Submission.Comment(treeID)
		if comment == nil {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		comments = []pages.Comment{comment}
	} else {
		for _, c := range q.Submission.Comments {
			comments = append(comments, c)
		}
	}
	templateData := &pages.Submission{
		ItemID:       q.Submission.ItemID,
		Title:        q.Submission.Title,
		Url:          q.Submission.Url,
		SubmittedAt:  q.Submission.SubmittedAt,
		Submitter:    q.Submission.Submitter,
		VoteCount:    q.Submission.VoteCount,
		CommentCount: q.Submission.CommentCount,
		Comments:     comments,
	}
	pages.ItemPage("/item", templateData, pageData).Render(w)
}
