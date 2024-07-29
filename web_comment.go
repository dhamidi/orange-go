package main

import (
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoComment(w http.ResponseWriter, req *http.Request) {
	currentUser := web.CurrentUser(req)
	itemID := req.FormValue("item_id")
	if currentUser == nil {
		web.LogInFirst(w, req)
		return
	}

	if req.Method == "GET" && isHX(req) {
		pages.CommentForm(req.FormValue("item_id"), pages.NewFormState()).Render(w)
		return
	}
	if req.Method != "POST" {
		back := req.Header.Get("Referer")
		http.Redirect(w, req, back, http.StatusSeeOther)
		return
	}

	cmd := &PostComment{
		ParentID: NewTreeID(itemID),
		Author:   currentUser.Username,
		PostedAt: web.CurrentTime(),
		Content:  req.FormValue("content"),
	}

	state := pages.NewFormState()
	if err := web.app.HandleCommand(cmd); err != nil {
		state.SetValue("content", req.FormValue("content"))
		state.AddError("content", err.Error())
		if isHX(req) {
			pages.CommentForm(itemID, state).Render(w)
			return
		}
	}

	w.Header().Set("HX-Redirect", req.Header.Get("Referer"))
	w.WriteHeader(http.StatusNoContent)
}
