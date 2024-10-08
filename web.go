package main

import (
	"context"
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
		fileServer:   http.FileServer(http.FS(staticFiles)),
		fs:           staticFiles,
		cacheForever: IMMUTABLE_ASSETS,
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
	routes.HandleFunc("/admin/a/unhide-submission", web.AdminOnly(web.DoUnhideSubmission))
	routes.HandleFunc("/admin/a/hide-submission", web.AdminOnly(web.DoHideSubmission))
	routes.HandleFunc("/admin/a/unhide-comment", web.AdminOnly(web.DoUnhideComment))
	routes.HandleFunc("/admin/a/hide-comment", web.AdminOnly(web.DoHideComment))
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
	fileServer   http.Handler
	fs           fs.FS
	cacheForever []string
}

func (z *WithGzipFS) shouldCache(path string) bool {
	for _, p := range z.cacheForever {
		if strings.HasSuffix(path, p) {
			return true
		}
	}
	return false
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
	}
	if z.shouldCache(req.URL.Path) {
		w.Header().Set("Cache-Control", "max-age=31536000, immutable")
	}
	z.fileServer.ServeHTTP(w, req)
}

func (web *WebApp) AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		currentUser := web.CurrentUser(req)
		if currentUser == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		getUserRoles := &Request{
			Headers:    Dict{"Name": "GetUserRoles", "Kind": "query"},
			Parameters: &url.Values{"username": []string{currentUser.Username}},
		}
		result, err := web.shell.Do(context.Background(), getUserRoles)
		userRoles := result.([]UserRole)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		if slices.Contains(userRoles, UserRoleAdmin) {
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
	stylesheet := "/s/main.css"
	for _, asset := range IMMUTABLE_ASSETS {
		if strings.HasPrefix(asset, "main.") && strings.HasSuffix(asset, ".css.gz") {
			stylesheet = asset[:len(asset)-len(".gz")]
		}
	}

	pageData := &pages.PageData{
		Stylesheet:  stylesheet,
		CurrentUser: nil,
		IsAdmin:     false,
		FormState:   pages.NewFormState(),
		MainOnly:    isHX(req),
	}
	if currentUser != nil {
		pageData.CurrentUser = &pages.User{Username: currentUser.Username}
		// fetching roles is on a best-effort basis
		getUserRoles := &Request{
			Headers:    Dict{"Name": "GetUserRoles", "Kind": "query"},
			Parameters: &url.Values{"username": []string{currentUser.Username}},
		}
		roles, err := web.shell.Do(req.Context(), getUserRoles)
		if err == nil && slices.Contains(roles.([]UserRole), UserRoleAdmin) {
			pageData.IsAdmin = true
		}
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
