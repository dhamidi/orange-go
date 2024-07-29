package main

import (
	"errors"
	"net/http"
	"net/url"
	"orange/pages"
)

func (web *WebApp) PageSubmit(w http.ResponseWriter, req *http.Request) {
	currentUser := web.CurrentUser(req)
	switch req.Method {
	case "GET":
		if currentUser == nil {
			http.Redirect(w, req, "/login?back_to="+url.QueryEscape("/submit"), http.StatusSeeOther)
			return
		}
		form := pages.NewFormState()
		pages.SubmitPage(req.URL.Path, form, web.PageData(req)).Render(w)
	case "POST":
		web.handleSubmission(w, req, currentUser)
	}
}

func (web *WebApp) handleSubmission(w http.ResponseWriter, req *http.Request, currentUser *User) {
	now := web.CurrentTime()
	pageData := web.PageData(req)
	submitLink := &PostLink{
		ItemID:      web.ItemIDGenerator(),
		Submitter:   currentUser.Username,
		SubmittedAt: now,
		Title:       req.FormValue("title"),
		Url:         req.FormValue("url"),
	}

	err := web.app.HandleCommand(submitLink)
	form := pages.NewFormState()
	form.SetValue("title", submitLink.Title)
	form.SetValue("url", submitLink.Url)
	if errors.Is(err, ErrEmptyTitle) {
		form.AddError("title", ErrEmptyTitle.Error())
	}
	if errors.Is(err, ErrEmptyUrl) {
		form.AddError("url", ErrEmptyUrl.Error())
	}
	if errors.Is(err, ErrMalformedURL) {
		form.AddError("url", "Only http and https URLs are supported")
	}

	if form.HasErrors() {
		pages.SubmitPage(req.URL.Path, form, pageData).Render(w)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
