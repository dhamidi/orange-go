package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func signup(app *App, params []string) {
	if len(params) < 2 {
		fmt.Printf("usage: signup <username> <password>\n")
		return
	}
	username, password := params[0], params[1]

	passwordHash, err := HashPassword(password)
	if err != nil {
		fmt.Printf("failed to hash password: %s\n", err)
		return
	}
	signup := &SignUpUser{
		Username:     username,
		PasswordHash: passwordHash.String(),
		CreatedAt:    time.Now(),
	}
	if err := app.HandleCommand(signup); err != nil {
		fmt.Printf("failed to sign up user: %s\n", err)
	}
}

func login(app *App, params []string) {
	if len(params) < 2 {
		fmt.Printf("usage: log-in <username> <password>\n")
		return
	}

	username, password := params[0], params[1]
	q := NewFindUserPasswordHash(username, password)
	if err := app.HandleQuery(q); err != nil {
		fmt.Printf("failed to hash password: %s\n", err)
		return
	}
	login := &LogInUser{
		Username:     username,
		PasswordHash: q.PasswordHash.String(),
		AttemptedAt:  time.Now(),
		SessionID:    uuid.NewString(),
	}
	if err := app.HandleCommand(login); err != nil {
		fmt.Printf("failed to log in user: %s\n", err)
		return
	}
	fmt.Printf("%s\n", login.SessionID)
}

func findSession(app *App, sessionID string) {
	q := NewFindSessionQuery(sessionID)
	if err := app.HandleQuery(q); err != nil {
		fmt.Printf("failed to find session %q: %s\n", sessionID, err)
		return
	}
	fmt.Printf("%+v\n", q.Session)
}

func list(app *App, params []string) {
	after := 0
	if len(params) > 0 {
		i, err := strconv.Atoi(params[0])
		if err == nil {
			after = i
		}
	}
	log := app.Commands
	commands, err := log.After(after)
	if err != nil {
		fmt.Printf("failed to list commands: %s\n", err)
	}

	formatFieldValue := func(v reflect.Value) string {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsZero() {
				return "<nil>"
			} else {
				return fmt.Sprintf("%q", v.Elem())
			}
		default:
			return fmt.Sprintf("%q", v.Interface())
		}
	}
	formatPayload := func(command Command) string {
		c := reflect.ValueOf(command).Elem() // &PostLink{...} -> PostLink{...}
		t := c.Type()                        // PostLink{...} -> PostLink
		buffer := bytes.NewBufferString("")
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fmt.Fprintf(buffer, " %s=%s", field.Name, formatFieldValue(c.Field(i)))
		}
		return buffer.String()
	}

	for command := range commands {
		fmt.Printf("%3d %20s %s\n", command.ID, command.Message.CommandName(), formatPayload(command.Message))
	}
}

func upvote(app *App, params []string) {
	if len(params) < 2 {
		fmt.Printf("usage: upvote <session-id> <item-id>\n")
		return
	}
	sessionID, itemID := params[0], params[1]
	q := NewFindSessionQuery(sessionID)
	if err := app.HandleQuery(q); err != nil {
		fmt.Printf("failed to find session %q: %s\n", sessionID, err)
		return
	}
	if q.Session == nil {
		fmt.Printf("You are not logged in\n")
		return
	}
	submit := &UpvoteSubmission{
		ItemID:  itemID,
		Voter:   q.Session.Username,
		VotedAt: time.Now(),
	}
	if err := app.HandleCommand(submit); err != nil {
		fmt.Printf("failed to upvote: %s\n", err)
	}
}

func submit(app *App, params []string) {
	if len(params) < 3 {
		fmt.Printf("usage: submit <session-id> <title> <url>\n")
		return
	}
	sessionID, title, url := params[0], params[1], params[2]
	q := NewFindSessionQuery(sessionID)
	if err := app.HandleQuery(q); err != nil {
		fmt.Printf("failed to find session %q: %s\n", sessionID, err)
		return
	}
	if q.Session == nil {
		fmt.Printf("You are not logged in\n")
		return
	}
	submit := &PostLink{
		ItemID:      uuid.NewString(),
		Submitter:   q.Session.Username,
		Title:       title,
		Url:         url,
		SubmittedAt: time.Now(),
	}
	if err := app.HandleCommand(submit); err != nil {
		fmt.Printf("failed to submit content: %s\n", err)
	}
}

func comment(app *App, params []string) {
	if len(params) < 3 {
		fmt.Printf("usage: comment <session-id> <item-id> <text>\n")
		return
	}
	sessionID, itemID, text := params[0], params[1], params[2]
	q := NewFindSessionQuery(sessionID)
	if err := app.HandleQuery(q); err != nil {
		fmt.Printf("failed to find session %q: %s\n", sessionID, err)
		return
	}
	if q.Session == nil {
		fmt.Printf("You are not logged in\n")
		return
	}
	commentOn := &PostComment{
		ParentID: NewTreeID(itemID),
		Author:   q.Session.Username,
		Content:  text,
		PostedAt: time.Now(),
	}
	if err := app.HandleCommand(commentOn); err != nil {
		fmt.Printf("failed to post comment: %s\n", err)
	}
}

func main() {
	config := NewPlatformConfigFromEnv(os.Getenv)
	app, starters := HackerNews(config)
	before := time.Now()
	if err := app.Replay(true); err != nil {
		fmt.Printf("failed to replay commands: %s\n", err)
		os.Exit(1)
	}
	after := time.Now()
	subcommand := "serve"
	if len(os.Args) >= 2 {
		subcommand = os.Args[1]
	}

	switch subcommand {
	case "signup":
		signup(app, os.Args[2:])
	case "log-in":
		login(app, os.Args[2:])
	case "comment":
		comment(app, os.Args[2:])
	case "submit":
		submit(app, os.Args[2:])
	case "upvote":
		upvote(app, os.Args[2:])
	case "find-session":
		sessionID := ""
		if len(os.Args) > 2 {
			sessionID = os.Args[2]
		}
		findSession(app, sessionID)
	case "log":
		list(app, os.Args[2:])
	case "serve":
		web := NewWebApp(app)
		conninfo := ":8080"
		if len(os.Args) > 2 {
			conninfo = os.Args[2]
		}
		fmt.Printf("Replayed events in %s\n", after.Sub(before))
		for _, starter := range starters {
			starter.Start()
		}
		fmt.Printf("%s\n", http.ListenAndServe(conninfo, web))
	default:
		fmt.Printf("unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
