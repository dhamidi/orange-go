package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoUnhideComment(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	req.ParseForm()
	sessionID, _ := req.Cookie("session_id")
	itemID := req.Form.Get("itemID")
	req.Form.Set("sessionID", sessionID.Value)

	unhideComment := &Request{
		Headers:    Dict{"Name": "UnhideComment", "Kind": "command"},
		Parameters: req.Form,
	}
	_, err := web.shell.Do(req.Context(), unhideComment)
	if errors.Is(err, ErrSessionNotFound) {
		pages.UnhideCommentButton(itemID).Render(w)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.HideCommentButton(itemID).Render(w)
	return
}
