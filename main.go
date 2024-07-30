package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	config := NewPlatformConfigFromEnv(os.Getenv)
	app, starters := HackerNews(config)
	before := time.Now()
	if err := app.Replay(false); err != nil {
		fmt.Printf("failed to replay commands: %s\n", err)
		os.Exit(1)
	}
	after := time.Now()
	subcommand := "serve"
	if len(os.Args) >= 2 {
		subcommand = os.Args[1]
	}
	shell := NewDefaultShell(app)

	switch subcommand {
	case "set-default-username-policy":
		setUsernamePolicy := &ChangeUsernamePolicy{
			MinLength: 10,
			MaxLength: 100,
			Blacklist: []string{"guest"},
		}
		if err := app.HandleCommand(setUsernamePolicy); err != nil {
			fmt.Printf("failed to set username policy: %s\n", err)
		}
	case "signup":
		shell.Signup(os.Args[2:])
	case "log-in":
		shell.Login(os.Args[2:])
	case "comment":
		shell.Comment(os.Args[2:])
	case "submit":
		shell.Submit(os.Args[2:])
	case "upvote":
		shell.Upvote(os.Args[2:])
	case "find-session":
		sessionID := ""
		if len(os.Args) > 2 {
			sessionID = os.Args[2]
		}
		shell.FindSession(sessionID)
	case "log":
		shell.List(os.Args[2:])
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
