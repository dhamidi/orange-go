package main

import (
	"fmt"
	"net/http"
	"orange/pages"
	"time"

	"github.com/google/uuid"
)

type WebApp struct {
	app *App
	mux *http.ServeMux

	SessionIDGenerator func() string
	ItemIDGenerator    func() string
	CurrentTime        func() time.Time
}

func NewWebApp(app *App) *WebApp {
	web := &WebApp{
		app:                app,
		mux:                http.NewServeMux(),
		SessionIDGenerator: uuid.NewString,
		ItemIDGenerator:    uuid.NewString,
		CurrentTime:        time.Now,
	}

	web.registerRoutes()
	return web
}

func (web *WebApp) registerRoutes() {
	routes := web.mux
	routes.HandleFunc("/upvote", web.DoUpvote)
	routes.HandleFunc("/submit", web.PageSubmit)
	routes.HandleFunc("/logout", web.DoLogOut)
	routes.HandleFunc("/login", web.PageLogin)
	routes.HandleFunc("/", web.PageIndex)
}

func (web *WebApp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s %s %s\n", web.CurrentTime().UTC().Format(time.StampMilli), req.Method, req.URL.Path)
	web.app.Replay()
	w.Header().Set("X-T", web.CurrentTime().Format(time.RFC3339))
	web.mux.ServeHTTP(w, req)
}

func (web *WebApp) CurrentUser(req *http.Request) *User {
	sessionCookie, _ := req.Cookie("session_id")
	if sessionCookie == nil {
		return nil
	}

	q := NewFindUserBySessionIDQuery(sessionCookie.Value)
	if err := web.app.HandleQuery(q); err != nil {
		return nil
	}
	return q.User
}

func (web *WebApp) PageData(req *http.Request) *pages.PageData {
	currentUser := web.CurrentUser(req)
	pageData := &pages.PageData{
		CurrentUser: nil,
		FormState:   pages.NewFormState(),
	}
	if currentUser != nil {
		pageData.CurrentUser = &pages.User{Username: currentUser.Username}
	}
	return pageData
}
