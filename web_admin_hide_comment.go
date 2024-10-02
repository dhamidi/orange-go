package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoHideComment(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	req.ParseForm()
	sessionID, _ := req.Cookie("session_id")
	itemID := req.Form.Get("itemID")
	req.Form.Set("sessionID", sessionID.Value)

	hideComment := &Request{
		Headers:    Dict{"Name": "HideComment", "Kind": "command"},
		Parameters: req.Form,
	}
	_, err := web.shell.Do(req.Context(), hideComment)
	if errors.Is(err, ErrSessionNotFound) {
		pages.HideCommentButton(itemID).Render(w)
		return
	}

	if err != nil {
		web.logger.Printf("error hiding comment: %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.UnhideCommentButton(itemID).Render(w)
	return
}
