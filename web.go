package main

import (
	"errors"
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
	routes.HandleFunc("/submit", web.PageSubmit)
	routes.HandleFunc("/logout", web.DoLogOut)
	routes.HandleFunc("/login", web.PageLogin)
	routes.HandleFunc("/", web.PageIndex)
}

func (web *WebApp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s %s %s\n", web.CurrentTime().UTC().Format(time.StampMilli), req.Method, req.URL.Path)
	web.app.Replay()
	web.mux.ServeHTTP(w, req)
}

func (web *WebApp) PageData(req *http.Request) *pages.PageData {
	currentUser := web.CurrentUser(req)
	pageData := &pages.PageData{
		CurrentUser: nil,
	}
	if currentUser != nil {
		pageData.CurrentUser = &pages.User{Username: currentUser.Username}
	}
	return pageData
}

func (web *WebApp) DoLogOut(w http.ResponseWriter, req *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: ""})
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (web *WebApp) PageSubmit(w http.ResponseWriter, req *http.Request) {
	currentUser := web.CurrentUser(req)
	switch req.Method {
	case "GET":
		if currentUser == nil {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}
		form := pages.NewFormState()
		pages.SubmitPage(req.URL.Path, form, web.PageData(req)).Render(w)
	case "POST":
		web.handleSubmission(w, req, currentUser)
	}
}

func (web *WebApp) handleSubmission(w http.ResponseWriter, req *http.Request, currentUser *User) {
	now := web.CurrentTime()
	pageData := web.PageData(req)
	submitLink := &PostLink{
		ItemID:      web.ItemIDGenerator(),
		Submitter:   currentUser.Username,
		SubmittedAt: now,
		Title:       req.FormValue("title"),
		Url:         req.FormValue("url"),
	}

	err := web.app.HandleCommand(submitLink)
	form := pages.NewFormState()
	form.SetValue("title", submitLink.Title)
	form.SetValue("url", submitLink.Url)
	if errors.Is(err, ErrEmptyTitle) {
		form.AddError("title", ErrEmptyTitle.Error())
	}
	if errors.Is(err, ErrEmptyUrl) {
		form.AddError("url", ErrEmptyUrl.Error())
	}
	if errors.Is(err, ErrMalformedURL) {
		form.AddError("url", "Only http and https URLs are supported")
	}

	if form.HasErrors() {
		pages.SubmitPage(req.URL.Path, form, pageData).Render(w)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (web *WebApp) PageLogin(w http.ResponseWriter, req *http.Request) {
	currentUser := web.CurrentUser(req)
	switch req.Method {
	case "GET":
		if currentUser != nil {
			http.Redirect(w, req, "/", http.StatusSeeOther)
			return
		}
		pages.LoginPage(req.URL.Path, nil).Render(w)
	case "POST":
		web.handleLogIn(w, req)
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

func (web *WebApp) handleLogIn(w http.ResponseWriter, req *http.Request) {
	sessionID := web.SessionIDGenerator()
	now := web.CurrentTime()
	passwordHash, err := HashPassword(req.FormValue("password"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	logIn := &LogInUser{
		Username:     req.FormValue("username"),
		PasswordHash: passwordHash.String(),
		AttemptedAt:  now,
		SessionID:    sessionID,
	}

	if err := web.app.HandleCommand(logIn); err != nil && err == ErrInvalidCredentials {
		http.Redirect(w, req, "/login?error=invalid-credentials", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
	})
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (web *WebApp) PageIndex(w http.ResponseWriter, req *http.Request) {
	q := NewFrontpageQuery()
	if err := web.app.HandleQuery(q); err != nil {
		fmt.Printf("Failed to load frontpage: %s\n", err)
		http.Error(w, "failed to load front page", http.StatusInternalServerError)
		return
	}
	templateData := []*pages.Submission{}
	for _, submission := range q.Submissions {
		title := submission.Title
		imageURL := (*string)(nil)
		if submission.Preview != nil {
			if submission.Preview.Title != nil {
				if len(*submission.Preview.Title) > len(submission.Title) {
					title = *submission.Preview.Title
				}
			}
			imageURL = submission.Preview.ImageURL
		}
		templateData = append(templateData, &pages.Submission{
			ItemID:      submission.ItemID,
			Title:       title,
			ImageURL:    imageURL,
			Url:         submission.Url,
			SubmittedAt: submission.SubmittedAt,
			Submitter:   submission.Submitter,
		})
	}
	if req.Header.Get("HX-Request") != "" {
		pages.SubmissionList(templateData).Render(w)
		return
	}
	_ = pages.IndexPage(req.URL.Path, templateData, web.PageData(req)).Render(w)
}
