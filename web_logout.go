package main

import "net/http"

func (web *WebApp) DoLogOut(w http.ResponseWriter, req *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: ""})
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
