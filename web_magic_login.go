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
		email := req.FormValue("email")
		params := url.Values{}
		params.Set("email", email)
		// ignore errors, as we don't want to leak if an email is registered
		if _, err := web.shell.RequestMagicLinkLogin(params); err != nil {
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

	sessionID, err := web.shell.LoginWithMagicLink(params)
	if err != nil {
		pages.ForbiddenMagicPage(req.URL.Path).Render(w)
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
