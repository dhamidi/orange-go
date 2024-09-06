package main

import (
	"net/http"
	"net/url"
	"orange/pages"
)

func (web *WebApp) PageRequestLoginWithMagic(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		pages.RequestMagicPage(req.URL.Path, web.PageData(req)).Render(w)
		return
	case "POST":
		req.ParseForm()
		requestLoginWithMagic := &Request{
			Headers:    Dict{"Name": "RequestLoginWithMagic", "Kind": "command"},
			Parameters: req.Form,
		}
		// ignore errors, as we don't want to leak if an email is registered
		if _, err := web.shell.Do(req.Context(), requestLoginWithMagic); err != nil {
			web.logger.Printf("failed to request magic link: %v", err)
		}

		if isHX(req) {
			pages.MagicSentPage().Render(w)
		}
	}
}

func (web *WebApp) PageLoginWithMagic(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	params := url.Values{}
	params.Set("magic", req.PathValue("magic"))

	sessionID := web.SessionIDGenerator()
	params.Set("sessionID", sessionID)
	logInWithMagic := &Request{
		Headers:    Dict{"Name": "LogInWithMagic", "Kind": "command"},
		Parameters: params,
	}
	_, err := web.shell.Do(req.Context(), logInWithMagic)
	if err != nil {
		pageData := web.PageData(req)
		pages.ForbiddenMagicPage(req.URL.Path, pageData).Render(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
