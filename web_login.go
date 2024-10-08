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
	req.ParseForm()
	sessionID := web.SessionIDGenerator()
	req.Form.Set("sessionID", sessionID)
	_, err := web.shell.Do(req.Context(), &Request{
		Headers:    Dict{"Name": "LogIn", "Kind": "command"},
		Parameters: req.Form,
	})
	if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrPasswordMismatch) {
		query := url.Values{}
		query.Set("error", "invalid-credentials")
		query.Set("username", req.FormValue("username"))
		http.Redirect(w, req, "/login?"+query.Encode(), http.StatusSeeOther)
		return
	}
	if err != nil {
		web.logger.Printf("failed to log in: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	})
	backTo := req.FormValue("back_to")
	if backTo == "" {
		backTo = "/"
	}
	http.Redirect(w, req, backTo, http.StatusSeeOther)
}
