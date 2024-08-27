package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoHideSubmission(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	req.ParseForm()
	sessionID, _ := req.Cookie("session_id")
	itemID := req.Form.Get("itemID")
	req.Form.Set("sessionID", sessionID.Value)

	err := web.shell.HideSubmission(req.Form)
	if errors.Is(err, ErrSessionNotFound) {
		pages.HideSubmissionButton(itemID, false).Render(w)
		return
	}

	if err == ErrAlreadyVoted {
		pages.HiddenIcon().Render(w)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.HiddenIcon().Render(w)
	return
}
