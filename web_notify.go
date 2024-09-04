package main

import (
	"fmt"
	"net/http"
	"time"
)

type Notification struct {
	Actor  string
	Entity string
	Action string
}

func (n *Notification) String() string {
	return fmt.Sprintf("%s %s %s", n.Actor, n.Action, n.Entity)
}

func NewNotificationFromCommand(cmd Command, app *App) (*Notification, bool) {
	switch cmd := cmd.(type) {
	case *PostComment:
		submissionID := cmd.ParentID[0]
		q := NewFindSubmission(submissionID)
		if err := app.HandleQuery(q); err != nil {
			return nil, false
		}
		return &Notification{
			Actor:  cmd.Author,
			Entity: q.Submission.Title,
			Action: "commented on",
		}, true
	default:
		return nil, false
	}
}

func (web *WebApp) DoNotify(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	commands, err := web.app.Commands.After(0)
	if err != nil {
		http.Error(w, "failed to subscribe", http.StatusInternalServerError)
		return
	}
	lastSeen := 0
	for cmd := range commands {
		lastSeen = cmd.ID
	}
	web.logger.Printf("notify(%d): from %s", lastSeen, req.RemoteAddr)
	everySecond := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-req.Context().Done():
			web.logger.Printf("notify(%d): from %s: done", lastSeen, req.RemoteAddr)
			return
		case <-everySecond.C:
			newCommands, err := web.app.Commands.After(lastSeen)
			if err != nil {
				continue
			}
			for cmd := range newCommands {
				notification, ok := NewNotificationFromCommand(cmd.Message, web.app)
				if ok {
					fmt.Fprintf(w, "event: notify\n")
					fmt.Fprintf(w, "data: %s\n\n", notification)
					w.(http.Flusher).Flush()
				}
				lastSeen = cmd.ID
			}
		}
	}
}
