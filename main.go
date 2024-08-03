package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func p(i int) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return ""
}

func pv(i int, key string, values *url.Values) {
	if len(os.Args) > i {
		values.Set(key, os.Args[i])
	}
}

func run(err error, usage ...string) {
	if err != nil {
		if len(usage) > 0 {
			fmt.Printf("usage %s\n", strings.Join(usage, " "))
		}
		fmt.Printf("failed: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	config := NewPlatformConfigFromEnv(os.Getenv)
	app, starters := HackerNews(config)
	before := time.Now()
	if err := app.Replay(config.SkipErrorsDuringReplay); err != nil {
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
			MinLength: 5,
			MaxLength: 100,
			Blacklist: []string{"guest"},
		}
		if err := app.HandleCommand(setUsernamePolicy); err != nil {
			fmt.Printf("failed to set username policy: %s\n", err)
		}
	case "signup":
		values := url.Values{}
		values.Set("username", p(2))
		values.Set("password", p(3))
		run(shell.Signup(values), "signup <username> <password>")
	case "link-email":
		values := url.Values{}
		values.Set("username", p(2))
		values.Set("email", p(3))
		run(shell.LinkVerifiedEmailToUser(values), "link-email <username> <email>")
	case "log-in":
		values := url.Values{}
		values.Set("username", p(2))
		values.Set("password", p(3))
		sessionID, err := shell.Login(values)
		run(err, "login <username> <password>")
		fmt.Printf("%s\n", sessionID)
	case "comment":
		shell.Comment(os.Args[2:])
	case "submit":
		shell.Submit(os.Args[2:])
	case "upvote":
		shell.Upvote(os.Args[2:])
	case "find-session":
		values := url.Values{}
		values.Set("sessionID", p(2))
		session, err := shell.FindSession(values)
		run(err, "find-session <session-id>")
		fmt.Printf("%s", session.ID)
	case "log":
		values := url.Values{}
		pv(2, "after", &values)
		run(shell.List(values, os.Stdout))
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
