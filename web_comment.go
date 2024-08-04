package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoComment(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if req.Method == "GET" && isHX(req) {
		pages.CommentForm(req.FormValue("itemID"), pages.NewFormState()).Render(w)
		return
	}
	if req.Method != "POST" {
		back := req.Header.Get("Referer")
		http.Redirect(w, req, back, http.StatusSeeOther)
		return
	}

	sessionID, _ := req.Cookie("session_id")
	itemID := req.FormValue("itemID")
	req.Form.Set("sessionID", sessionID.Value)

	err := web.shell.Comment(req.Form)
	if errors.Is(err, ErrSessionNotFound) {
		web.LogInFirst(w, req)
		return
	}
	if err != nil {
		state := pages.NewFormState()
		state.SetValue("text", req.FormValue("text"))
		state.AddError("text", err.Error())
		if isHX(req) {
			pages.CommentForm(itemID, state).Render(w)
			return
		}
	}

	w.Header().Set("HX-Redirect", req.Header.Get("Referer"))
	w.WriteHeader(http.StatusNoContent)
}
