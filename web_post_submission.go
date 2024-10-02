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
		web.handleSubmission(w, req)
	}
}

func (web *WebApp) handleSubmission(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	pageData := web.PageData(req)
	sessionID, _ := req.Cookie("session_id")
	req.Form.Set("sessionID", sessionID.Value)
	req.Form.Set("itemID", web.ItemIDGenerator())

	form := pages.NewFormState()
	form.SetValue("title", req.Form.Get("title"))
	form.SetValue("url", req.Form.Get("url"))

	if !LinkIsLive(form.Values["url"]) {
		form.AddError("url", "URL is not reachable")
		pages.SubmitPage(req.URL.Path, form, pageData).Render(w)
		return
	}

	submit := &Request{
		Headers:    Dict{"Name": "PostLink", "Kind": "command"},
		Parameters: req.Form,
	}
	_, err := web.shell.Do(req.Context(), submit)
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
