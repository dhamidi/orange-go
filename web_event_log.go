package main

import (
	"fmt"
	"iter"
	"net/http"
	"orange/pages"
	"slices"
	"strconv"
)

func (web *WebApp) PageEventLog(w http.ResponseWriter, req *http.Request) {
	pageData := web.PageData(req)
	if user := pageData.Username(); user == nil {
		http.Error(w, "not authorized", http.StatusForbidden)
		return
	}
	allCommands := []*pages.EventLogEntry{}

	nStr := req.FormValue("n")
	n := 10
	if nInt, err := strconv.Atoi(nStr); nStr != "" && err == nil {
		n = nInt
	}
	var (
		commandsAfter iter.Seq[*PersistedCommand]
		length        int
		after         int
		err           error
	)

	length, err = web.app.Commands.Length()
	if err != nil {
		goto internalError
	}

	if a, err := strconv.Atoi(req.FormValue("after")); err == nil {
		after = a
	} else {
		after = length - n
	}

	commandsAfter, err = web.app.Commands.After(after)
	if err != nil {
		goto internalError
	}

	for command := range commandsAfter {
		if n == 0 {
			break
		} else {
			n--
		}
		allCommands = append(allCommands, &pages.EventLogEntry{
			ID:      fmt.Sprintf("%d", command.ID),
			Kind:    command.Message.CommandName(),
			Payload: command.Message,
		})
	}
	slices.Reverse(allCommands)
	pages.EventLogPage(req.URL.Path, allCommands, pageData).Render(w)
	return
internalError:
	web.logger.Printf("failed to load commands: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
