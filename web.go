package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"orange/pages"
	"os"
	"time"

	"embed"

	"github.com/google/uuid"
)

//go:embed static
var embeddedStaticFiles embed.FS

type WebApp struct {
	app    *App
	mux    *http.ServeMux
	logger *log.Logger

	SessionIDGenerator func() string
	ItemIDGenerator    func() string
	CurrentTime        func() time.Time
}

func NewWebApp(app *App) *WebApp {
	web := &WebApp{
		app:                app,
		mux:                http.NewServeMux(),
		logger:             log.New(os.Stdout, "[web]", log.LstdFlags),
		SessionIDGenerator: uuid.NewString,
		ItemIDGenerator:    uuid.NewString,
		CurrentTime:        time.Now,
	}

	web.registerRoutes()
	return web
}

func (web *WebApp) registerRoutes() {
	staticFiles, _ := fs.Sub(embeddedStaticFiles, "static")
	routes := web.mux
	routes.HandleFunc("/notify", web.DoNotify)
	routes.HandleFunc("/comment", web.DoComment)
	routes.HandleFunc("/item", web.PageItem)
	routes.HandleFunc("/upvote", web.DoUpvote)
	routes.HandleFunc("/submit", web.PageSubmit)
	routes.HandleFunc("/logout", web.DoLogOut)
	routes.HandleFunc("/login", web.PageLogin)
	routes.HandleFunc("/admin/events", web.PageEventLog)
	routes.Handle("/favicon.ico", http.FileServer(http.FS(staticFiles)))
	routes.HandleFunc("/", web.PageIndex)
}

func (web *WebApp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s %s %s\n", web.CurrentTime().UTC().Format(time.StampMilli), req.Method, req.URL.Path)
	web.app.Replay(true)
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
	backToRaw := ""
	if backTo := req.FormValue("back_to"); backTo != "" {
		backToRaw = backTo
	} else if req.Header.Get("X-Back-To") != "" {
		backToRaw = req.Header.Get("X-Back-To")
	}
	backToURL, err := url.Parse(backToRaw)
	if err == nil {
		pageData.BackTo = backToURL
	}
	return pageData
}

func isHX(req *http.Request) bool {
	return req.Header.Get("HX-Request") != ""
}

func (web *WebApp) LogInFirst(w http.ResponseWriter, req *http.Request) {
	referer := &url.URL{Path: req.URL.Path, RawQuery: req.URL.Query().Encode()}
	if currentURL := req.Header.Get("HX-Current-Url"); currentURL != "" {
		backTo, _ := url.Parse(currentURL)
		referer.Path = backTo.Path
		referer.RawQuery = backTo.RawQuery
	}
	redirectTo := &url.URL{
		Path:     "/login",
		RawQuery: url.Values{"back_to": []string{referer.String()}}.Encode(),
	}

	if isHX(req) {
		w.Header().Add("HX-Redirect", redirectTo.String())
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Redirect(w, req, req.URL.String(), http.StatusSeeOther)
	}
}
