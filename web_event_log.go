package main

import (
	"fmt"
	"net/http"
	"orange/pages"
)

func (web *WebApp) PageEventLog(w http.ResponseWriter, req *http.Request) {
	pageData := web.PageData(req)
	if user := pageData.Username(); user == nil {
		http.Error(w, "not authorized", http.StatusForbidden)
		return
	}
	allCommands := []*pages.EventLogEntry{}

	commandsAfter, err := web.app.Commands.After(0)
	if err != nil {
		web.logger.Printf("failed to load commands: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	for command := range commandsAfter {
		allCommands = append(allCommands, &pages.EventLogEntry{
			ID:      fmt.Sprintf("%d", command.ID),
			Kind:    command.Message.CommandName(),
			Payload: command.Message,
		})
	}
	pages.EventLogPage(req.URL.Path, allCommands, pageData).Render(w)
}
