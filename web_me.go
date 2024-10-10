package main

import (
	"errors"
	"fmt"
	"net/http"
	"orange/pages"
)

func (web *WebApp) PageMe(w http.ResponseWriter, req *http.Request) {
	currentUser := web.CurrentUser(req)
	if currentUser == nil {
		web.LogInFirst(w, req)
		return
	}

	if req.Method == "POST" {
		web.pageMeUpdateSettings(w, req)
		return
	}

	templateData := &pages.AccountDetails{
		Username:                currentUser.Username,
		Email:                   currentUser.VerifiedEmail,
		EmailVerified:           true,
		SubscribedToSubmissions: false,
		SubscribedToReplies:     false,
	}
	q := NewSubscriptionSettingsForUserQuery(currentUser.Username)
	if err := web.app.HandleQuery(q); err != nil && !errors.Is(err, ErrSubscriptionSettingsNotFound) {
		web.logger.Printf("Failed to retrieve subscription settings for %q: %s", currentUser.Username, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if err == nil {
		templateData.SubscribedToReplies = q.Settings.HasScope(SUBSCRIPTION_SCOPE_REPLIES)
		templateData.SubscribedToSubmissions = q.Settings.HasScope(SUBSCRIPTION_SCOPE_SUBMISSIONS)
	}
	_ = pages.MePage(templateData, web.PageData((req))).Render(w)
}

func (web *WebApp) pageMeUpdateSettings(w http.ResponseWriter, req *http.Request) {
	sessionID, _ := req.Cookie("session_id")
	if sessionID == nil {
		web.LogInFirst(w, req)
		return
	}

	paramsEnable := Dict{}
	paramsDisable := Dict{}
	enable := []string{}
	disable := []string{}
	updateEnable := &Request{
		Headers:    Dict{"sessionID": sessionID.Value, "Kind": "command", "Name": "EnableSubscriptions"},
		Parameters: paramsEnable,
	}
	updateDisable := &Request{
		Headers:    Dict{"sessionID": sessionID.Value, "Kind": "command", "Name": "DisableSubscriptions"},
		Parameters: paramsDisable,
	}
	if req.FormValue("subscribe_to_submissions") != "" {
		enable = append(enable, SUBSCRIPTION_SCOPE_SUBMISSIONS)
	} else {
		disable = append(disable, SUBSCRIPTION_SCOPE_SUBMISSIONS)
	}
	if req.FormValue("subscribe_to_replies") != "" {
		enable = append(enable, SUBSCRIPTION_SCOPE_REPLIES)
	} else {
		disable = append(disable, SUBSCRIPTION_SCOPE_REPLIES)
	}

	for i, scope := range enable {
		paramsEnable[fmt.Sprintf("scope[%d]", i)] = scope
	}
	for i, scope := range disable {
		paramsDisable[fmt.Sprintf("scope[%d]", i)] = scope
	}

	if len(paramsEnable) > 0 {
		if _, err := web.shell.Do(req.Context(), updateEnable); err != nil {
			web.logger.Printf("pageMeUpdateSettings: %s", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	if len(paramsDisable) > 0 {
		if _, err := web.shell.Do(req.Context(), updateDisable); err != nil {
			web.logger.Printf("pageMeUpdateSettings: %s", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, req, "/me", http.StatusSeeOther)
}
