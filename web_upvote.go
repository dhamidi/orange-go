package main

import (
	"net/http"
	"orange/pages"
)

func (web *WebApp) DoUpvote(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	currentUser := web.CurrentUser(req)
	itemID := req.FormValue("item_id")
	if currentUser == nil {
		pages.UpvoteButton(itemID).Render(w)
		return
	}
	cmd := &UpvoteSubmission{
		Voter:   currentUser.Username,
		VotedAt: web.CurrentTime(),
		ItemID:  itemID,
	}
	err := web.app.HandleCommand(cmd)
	if err == ErrAlreadyVoted {
		pages.VotedIcon().Render(w)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pages.VotedIcon().Render(w)
	return
}
