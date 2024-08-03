package main

import (
	"net/http"
	"orange/pages"
)

func (web *WebApp) PageMe(w http.ResponseWriter, req *http.Request) {
	currentUser := web.CurrentUser(req)
	if currentUser == nil {
		web.LogInFirst(w, req)
		return
	}

	templateData := &pages.AccountDetails{
		Username:      currentUser.Username,
		Email:         currentUser.VerifiedEmail,
		EmailVerified: true,
	}
	_ = pages.MePage(templateData, web.PageData((req))).Render(w)
}
