package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoComment(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	sessionID, _ := req.Cookie("session_id")
	if sessionID == nil || sessionID.Value == "" {
		web.LogInFirst(w, req)
		return
	}

	if req.Method == "GET" && isHX(req) {
		pages.CommentForm(req.FormValue("itemID"), pages.NewFormState()).Render(w)
		return
	}
	if req.Method != "POST" {
		back := req.Header.Get("Referer")
		http.Redirect(w, req, back, http.StatusSeeOther)
		return
	}

	itemID := req.FormValue("itemID")
	req.Form.Set("sessionID", sessionID.Value)

	comment := &Request{
		Headers:    Dict{"Name": "Comment", "Kind": "command"},
		Parameters: req.Form,
	}
	_, err := web.shell.Do(req.Context(), comment)
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
