package main

import (
	"fmt"
	"net/http"
	"os"
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
	passwordHash, err := HashPassword(password)
	if err != nil {
		fmt.Printf("failed to hash password: %s\n", err)
		return
	}
	login := &LogInUser{
		Username:     username,
		PasswordHash: passwordHash.String(),
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

func list(log CommandLog) {
	commands, err := log.After(0)
	if err != nil {
		fmt.Printf("failed to list commands: %s\n", err)
	}
	for command := range commands {
		fmt.Printf("%3d %20s %+v\n", command.ID, command.Message.CommandName(), command.Message)
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

func main() {
	config := NewPlatformConfigFromEnv(os.Getenv)
	app := HackerNews(config)
	if err := app.Replay(); err != nil {
		fmt.Printf("failed to replay commands: %s\n", err)
		os.Exit(1)
	}
	subcommand := "serve"
	if len(os.Args) >= 2 {
		subcommand = os.Args[1]
	}

	switch subcommand {
	case "signup":
		signup(app, os.Args[2:])
	case "log-in":
		login(app, os.Args[2:])
	case "submit":
		submit(app, os.Args[2:])
	case "find-session":
		sessionID := ""
		if len(os.Args) > 2 {
			sessionID = os.Args[2]
		}
		findSession(app, sessionID)
	case "log":
		list(app.Commands)
	case "serve":
		web := NewWebApp(app)
		conninfo := ":8080"
		if len(os.Args) > 2 {
			conninfo = os.Args[2]
		}
		fmt.Printf("%s\n", http.ListenAndServe(conninfo, web))
	default:
		fmt.Printf("unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
