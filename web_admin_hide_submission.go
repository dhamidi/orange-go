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

	hideSubmission := &Request{
		Headers:    Dict{"Name": "HideSubmission", "Kind": "command"},
		Parameters: req.Form,
	}
	_, err := web.shell.Do(req.Context(), hideSubmission)
	if errors.Is(err, ErrSessionNotFound) {
		pages.HideSubmissionButton(itemID).Render(w)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.UnhideSubmissionButton(itemID).Render(w)
	return
}
