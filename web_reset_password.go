package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) PageResetPassword(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		pageData := web.PageData(req)
		_ = pages.ResetPasswordPage("/reset-password", pageData).Render(w)
		return
	case "POST":
		pageData := web.PageData(req)
		form := pageData.FormState
		if req.FormValue("email") == "" {
			form.AddError("email", "Email is required")
			if isHX(req) {
				pages.ResetPasswordForm(form).Render(w)
				return
			}
			_ = pages.ResetPasswordPage("/reset-password", pageData).Render(w)
			return
		}
		requestPasswordReset := &Request{
			Headers:    Dict{"Name": "RequestPasswordReset", "Kind": "command"},
			Parameters: req.Form,
		}
		_, err := web.shell.Do(req.Context(), requestPasswordReset)
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			web.logger.Printf("error requesting password reset: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		if isHX(req) {
			pages.ResetPasswordSent().Render(w)
			return
		}
		_ = pages.ResetPasswordSentPage("/reset-password", pageData).Render(w)
		return
	}
}
