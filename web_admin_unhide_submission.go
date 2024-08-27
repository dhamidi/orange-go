package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoUnhideSubmission(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	req.ParseForm()
	sessionID, _ := req.Cookie("session_id")
	itemID := req.Form.Get("itemID")
	req.Form.Set("sessionID", sessionID.Value)

	err := web.shell.UnhideSubmission(req.Form)
	if errors.Is(err, ErrSessionNotFound) {
		pages.UnhideSubmissionButton(itemID).Render(w)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.HideSubmissionButton(itemID).Render(w)
	return
}
