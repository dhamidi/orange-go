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

func (web *WebApp) handleLogIn(w http.ResponseWriter, req *http.Request) {
	sessionID := web.SessionIDGenerator()
	now := web.CurrentTime()
	q := NewFindUserPasswordHash(req.FormValue("username"), req.FormValue("password"))
	query := url.Values{}
	query.Set("error", "invalid-credentials")
	query.Set("username", req.FormValue("username"))
	if err := web.app.HandleQuery(q); err != nil || q.PasswordHash == nil {
		http.Redirect(w, req, "/login?"+query.Encode(), http.StatusSeeOther)
		return
	}

	logIn := &LogInUser{
		Username:     req.FormValue("username"),
		PasswordHash: q.PasswordHash.String(),
		AttemptedAt:  now,
		SessionID:    sessionID,
	}

	err := web.app.HandleCommand(logIn)
	if errors.Is(err, ErrInvalidCredentials) {
		http.Redirect(w, req, "/login?"+query.Encode(), http.StatusSeeOther)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
