package main

import (
	"errors"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"net/url"
	"orange/pages"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"embed"

	"github.com/google/uuid"
)

//go:embed static
var embeddedStaticFiles embed.FS

type WebApp struct {
	app    *App
	shell  *Shell
	mux    *http.ServeMux
	logger *log.Logger

	SessionIDGenerator func() string
	ItemIDGenerator    func() string
	CurrentTime        func() time.Time
}

func NewWebApp(app *App, shell *Shell) *WebApp {
	web := &WebApp{
		app:                app,
		shell:              shell,
		mux:                http.NewServeMux(),
		logger:             log.New(os.Stdout, "[web] ", log.LstdFlags),
		SessionIDGenerator: uuid.NewString,
		ItemIDGenerator:    uuid.NewString,
		CurrentTime:        time.Now,
	}

	web.registerRoutes()
	return web
}

func (web *WebApp) registerRoutes() {
	staticFiles, _ := fs.Sub(embeddedStaticFiles, "static")
	staticFileServer := &WithGzipFS{
		fileServer: http.FileServer(http.FS(staticFiles)),
		fs:         staticFiles,
	}

	routes := web.mux
	routes.HandleFunc("/notify", web.DoNotify)
	routes.HandleFunc("/comment", web.DoComment)
	routes.HandleFunc("/item", web.PageItem)
	routes.HandleFunc("/upvote", web.DoUpvote)
	routes.HandleFunc("/submit", web.PageSubmit)
	routes.HandleFunc("/logout", web.DoLogOut)
	routes.HandleFunc("/login", web.PageLogin)
	routes.HandleFunc("/magic", web.PageRequestLoginWithMagic)
	routes.HandleFunc("/reset-password/{token}", web.PageResetPasswordToken)
	routes.HandleFunc("/reset-password", web.PageResetPassword)
	routes.HandleFunc("/login/{magic}", web.PageLoginWithMagic)
	routes.HandleFunc("/me", web.PageMe)
	routes.HandleFunc("/admin/events", web.AdminOnly(web.PageEventLog))
	routes.Handle("/favicon.ico", http.FileServer(http.FS(staticFiles)))
	routes.Handle("/s/", http.StripPrefix("/s/", staticFileServer))
	routes.HandleFunc("/", web.PageIndex)
}

func (web *WebApp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	web.logger.Printf("%s %s", req.Method, req.URL)
	web.app.Replay(true)
	w.Header().Set("X-T", web.CurrentTime().Format(time.RFC3339))
	web.mux.ServeHTTP(w, req)
}

type WithGzipFS struct {
	fileServer http.Handler
	fs         fs.FS
}

func (z *WithGzipFS) FileExists(path string) bool {
	_, err := z.fs.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

func (z *WithGzipFS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	uncompressed := req.URL.Path
	compressed := req.URL.Path + ".gz"
	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") && z.FileExists(compressed) {
		contentType := mime.TypeByExtension(filepath.Ext(uncompressed))
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", contentType)
		req.URL.Path = compressed
		z.fileServer.ServeHTTP(w, req)
	} else {
		z.fileServer.ServeHTTP(w, req)
	}
}

func (web *WebApp) AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		currentUser := web.CurrentUser(req)
		if currentUser == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		roles, err := web.shell.GetUserRoles(&url.Values{"username": []string{currentUser.Username}})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		if slices.Contains(roles, UserRoleAdmin) {
			next(w, req)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	}
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

	if referer.Path == req.URL.Path {
		referer.Path = "/"
	}

	redirectTo := &url.URL{
		Path:     "/login",
		RawQuery: url.Values{"back_to": []string{referer.String()}}.Encode(),
	}

	if isHX(req) {
		w.Header().Add("HX-Redirect", redirectTo.String())
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Redirect(w, req, redirectTo.String(), http.StatusSeeOther)
	}
}
