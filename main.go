package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kr/pretty"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Dict map[string]string

func (d Dict) Get(key string) string {
	return d[key]
}

func DictFromList(list []string) Dict {
	d := Dict{}
	for i := 0; i < len(list); i += 2 {
		d[list[i]] = list[i+1]
	}
	return d
}

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
	case "log":
		pv(2, "after", &values)
		run(shell.List(values, os.Stdout))
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
	case "do", "get":
		if len(os.Args) < 3 {
			fmt.Printf("usage: %s <command> <args...>\n", subcommand)
			os.Exit(1)
		}
		name := toCamelCase(os.Args[2])
		args := os.Args[3:]
		headers := Dict{"Kind": "command", "Name": name}

		if subcommand == "get" {
			headers["Kind"] = "query"
		}
		parameters := Dict{}
		for i := 0; i < len(args)-1; i += 2 {
			dest := parameters
			key := args[i]
			value := args[i+1]
			if strings.HasPrefix(key, "-") {
				key = key[1:]
				dest = headers
			}
			dest[key] = value
		}
		fmt.Printf("> headers: %#v\n> parameters: %#v\n", headers, parameters)
		req := &Request{Headers: headers, Parameters: parameters}
		result, err := shell.Do(context.Background(), req)
		fmt.Printf("< result: %# v\n", pretty.Formatter(result))
		fmt.Printf("< err: %s\n", err)
	default:
		fmt.Printf("unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

var titler = cases.Title(language.AmericanEnglish)

func toCamelCase(s string) string {
	components := strings.Split(s, "-")
	for i := range components {
		components[i] = titler.String(components[i])
	}
	return strings.Join(components, "")
}
