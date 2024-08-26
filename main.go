package main

import (
	"fmt"
	"log"
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

	values := url.Values{}

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
		pv(2, "username", &values)
		pv(3, "password", &values)
		run(shell.Signup(values), "signup <username> <password>")
	case "link-email":
		pv(2, "username", &values)
		pv(3, "email", &values)
		run(shell.LinkVerifiedEmailToUser(values), "link-email <username> <email>")
	case "log-in":
		pv(2, "username", &values)
		pv(3, "password", &values)
		sessionID, err := shell.Login(values)
		run(err, "login <username> <password>")
		fmt.Printf("%s\n", sessionID)
	case "request-password-reset":
		pv(2, "email", &values)
		token, err := shell.RequestPasswordReset(values)
		run(err, "request-password-reset <email>")
		fmt.Printf("%s\n", token)
	case "reset-password":
		pv(2, "token", &values)
		pv(3, "password", &values)
		run(shell.ResetPassword(values), "reset-password <token> <password>")
	case "request-magic-link":
		pv(2, "email", &values)
		magic, err := shell.RequestMagicLinkLogin(values)
		run(err, "request-magic-link <email>")
		fmt.Printf("%s\n", magic)
	case "log-in-with-magic":
		pv(2, "magic", &values)
		sessionID, err := shell.LoginWithMagicLink(values)
		run(err, "log-in-with-magic <magic>")
		fmt.Printf("%s\n", sessionID)
	case "hide-submission":
		pv(2, "sessionID", &values)
		pv(3, "itemID", &values)
		run(shell.HideSubmission(values), "hide-submission <session-id> <item-id>")
	case "unhide-submission":
		pv(2, "sessionID", &values)
		pv(3, "itemID", &values)
		run(shell.UnhideSubmission(values), "unhide-submission <session-id> <item-id>")
	case "comment":
		pv(2, "sessionID", &values)
		pv(3, "itemID", &values)
		pv(4, "text", &values)
		run(shell.Comment(values), "comment <session-id> <item-id> <text>")
	case "submit":
		pv(2, "sessionID", &values)
		pv(3, "title", &values)
		pv(4, "url", &values)
		run(shell.Submit(values), "submit <session-id> <title> <url>")
	case "upvote":
		pv(2, "sessionID", &values)
		pv(3, "itemID", &values)
		run(shell.Upvote(values), "upvote <session-id> <item-id>")
	case "find-session":
		values.Set("sessionID", p(2))
		session, err := shell.FindSession(values)
		run(err, "find-session <session-id>")
		fmt.Printf("%s", session.ID)
  case "frontpage":
		values.Set("sessionID", p(2))
		submissions, err := shell.GetFrontpage(values)
    run(err, "frontpage <session-id>")
    for _, s := range submissions {
      status := "active"
      if s.Hidden { status = "hidden" }
      fmt.Printf("%s %.2f %-8s %s\n", s.ItemID, s.Score, status, s.Url)
    }
	case "log":
		pv(2, "after", &values)
		run(shell.List(values, os.Stdout))
	case "queue-email":
		pv(2, "email", &values)
		pv(3, "subject", &values)
		cmd := &QueueEmail{
			InternalID:   shell.NewID(),
			Recipients:   values.Get("email"),
			Subject:      values.Get("subject"),
			TemplateName: "test",
			TemplateData: map[string]interface{}{"name": "test"},
		}
		run(app.HandleCommand(cmd), "queue-email <email> <subject>")
	case "send-welcome-email":
		pv(2, "username", &values)
		pv(3, "email", &values)
		cmd := &QueueEmail{
			InternalID:   shell.NewID(),
			Recipients:   values.Get("email"),
			Subject:      "",
			TemplateName: "welcome",
			TemplateData: map[string]interface{}{
				"name":      values.Get("username"),
				"username":  values.Get("username"),
				"login_url": "https://orange.decode.ee/login",
			},
		}
		run(app.HandleCommand(cmd), "send-welcome-email <username> <email>")
	case "unskip-commands":
		for i, arg := range os.Args[2:] {
			values.Set(fmt.Sprintf("id[%d]", i), arg)
		}
		run(shell.UnskipCommands(values), "unskip-commands <id>...")
	case "skip-commands":
		for i, arg := range os.Args[2:] {
			values.Set(fmt.Sprintf("id[%d]", i), arg)
		}
		run(shell.SkipCommands(values), "skip-commands <id>...")
	case "set-admin-users":
		for i, arg := range os.Args[2:] {
			values.Set(fmt.Sprintf("username[%d]", i), arg)
		}
		run(shell.SetAdminUsers(values), "set-admin-users <username>...")
	case "set-magic-domains":
		for i, arg := range os.Args[2:] {
			values.Set(fmt.Sprintf("domain[%d]", i), arg)
		}
		run(shell.SetMagicDomains(values), "set-magic-domains <username>...")
	case "serve":
		web := NewWebApp(app, shell)
		conninfo := ":8080"
		if len(os.Args) > 2 {
			conninfo = os.Args[2]
		}
		web.logger.Printf("Replayed events in %s\n", after.Sub(before))
		for _, starter := range starters {
			starter.Start()
		}
		go replServer(shell, log.New(os.Stdout, "[repl] ", log.LstdFlags))
		fmt.Printf("%s\n", http.ListenAndServe(conninfo, web))
	default:
		fmt.Printf("unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
