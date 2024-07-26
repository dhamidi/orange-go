package main

import (
	"fmt"
	"net/http"
	"orange/pages"
)

type WebApp struct {
	app *App
	mux *http.ServeMux
}

func NewWebApp(app *App) *WebApp {
	web := &WebApp{
		app: app,
		mux: http.NewServeMux(),
	}

	web.registerRoutes()
	return web
}

func (web *WebApp) registerRoutes() {
	routes := web.mux
	routes.HandleFunc("/", web.PageIndex)
}

func (web *WebApp) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	web.app.Replay()
	web.mux.ServeHTTP(w, req)
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
		templateData = append(templateData, &pages.Submission{
			Title: submission.Title,
			Url:   submission.Url,
		})
	}
	if req.Header.Get("HX-Request") != "" {
		pages.SubmissionList(templateData).Render(w)
		return
	}
	_ = pages.IndexPage(req.URL.Path, templateData).Render(w)
}
