package main

import (
	"errors"
	"net/http"
	"net/url"
	"orange/pages"
)

func (web *WebApp) PageLogin(w http.ResponseWriter, req *http.Request) {
	currentUser := web.CurrentUser(req)
	switch req.Method {
	case "GET":
		if currentUser != nil {
			http.Redirect(w, req, "/", http.StatusSeeOther)
			return
		}
		pageData := web.PageData(req)
		pageData.FormState.SetValue("username", req.FormValue("username"))
		if req.FormValue("error") == "invalid-credentials" {
			pageData.FormState.AddError("password", "Invalid credentials")
		}
		pages.LoginPage(req.URL.Path, pageData).Render(w)
	case "POST":
		web.handleLogIn(w, req)
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
		HttpOnly: true,
	})

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (web *WebApp) handleLogIn(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	sessionID, err := web.shell.Login(req.Form)
	if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrSessionNotFound) {
		query := url.Values{}
		query.Set("error", "invalid-credentials")
		query.Set("username", req.FormValue("username"))
		http.Redirect(w, req, "/login?"+query.Encode(), http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
	})
	backTo := req.FormValue("back_to")
	if backTo == "" {
		backTo = "/"
	}
	http.Redirect(w, req, backTo, http.StatusSeeOther)
}
