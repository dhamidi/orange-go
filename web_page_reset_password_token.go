package main

import (
	"errors"
	"net/http"
	"orange/pages"
)

func (web *WebApp) PageResetPasswordToken(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		pageData := web.PageData(req)
		form := pageData.FormState
		form.SetValue("token", req.PathValue("token"))
		_ = pages.ResetPasswordTokenPage("/reset-password", pageData).Render(w)
		return
	case "POST":
		pageData := web.PageData(req)
		form := pageData.FormState
		if req.FormValue("token") == "" {
			form.AddError("token", "Token is required")
		}
		if req.FormValue("password") == "" {
			form.AddError("password", "Password is required")
		}
		if form.HasErrors() {
			if isHX(req) {
				pages.ResetPasswordTokenForm(form).Render(w)
				return
			}
			_ = pages.ResetPasswordTokenPage("/reset-password", pageData).Render(w)
			return
		}
		err := web.shell.ResetPassword(req.Form)
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			web.logger.Printf("error resetting password: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		if isHX(req) {
			pages.ResetPasswordSuccess().Render(w)
			return
		}
		_ = pages.ResetPasswordSuccessPage("/reset-password", pageData).Render(w)
		return
	}
}
