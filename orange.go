package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"time"
)

type PlatformConfig struct {
	SkipErrorsDuringReplay bool
	ContentStore           *url.URL
	AuthStore              *url.URL
	CommandLog             *url.URL
}

func parseURL(u, field string) *url.URL {
	result, err := url.Parse(u)
	if err != nil {
		panic(fmt.Errorf("Error parsing URL %q for %s: %w", u, field, err))
	}
	return result
}

func toFilePath(u *url.URL) string {
	path := u.Path
	if path == "" {
		return path
	}
	if path[0] == '/' {
		return path[1:]
	}
	return path
}

func DefaultPlatformConfig() *PlatformConfig {
	return &PlatformConfig{
		SkipErrorsDuringReplay: false,
		ContentStore:           parseURL("memory://", "ContentStore"),
		AuthStore:              parseURL("memory://", "AuthStore"),
		CommandLog:             parseURL("file:///commands.db", "CommandLog"),
	}
}

func NewPlatformConfigForTest() *PlatformConfig {
	config := DefaultPlatformConfig()
	config.CommandLog = parseURL("memory://", "CommandLog")
	return config
}
func NewPlatformConfigFromEnv(getenv func(key string) string) *PlatformConfig {
	config := DefaultPlatformConfig()
	fields := []**url.URL{&config.ContentStore, &config.AuthStore, &config.CommandLog}
	names := []string{"CONTENT_STORE", "AUTH_STORE", "COMMAND_LOG"}
	for i := range fields {
		newURL := getenv("ORANGE_" + names[i])
		if newURL == "" {
			continue
		}
		*fields[i] = parseURL(newURL, names[i])
	}

	config.SkipErrorsDuringReplay = getenv("ORANGE_SKIP_ERRORS") == "true"

	return config
}

func (c *PlatformConfig) NewCommandLog() CommandLog {
	if c.CommandLog.Scheme == "file" {
		return NewFileCommandLog(toFilePath(c.CommandLog), DefaultSerializer)
	} else if c.CommandLog.Scheme == "memory" {
		return NewInMemoryCommandLog()
	} else {
		panic("Unsupported command log URL " + c.CommandLog.String())
	}
}

func (c *PlatformConfig) NewContentState() ContentState {
	if c.ContentStore.Scheme == "file" {
		return NewPersistentContentState(toFilePath(c.ContentStore))
	} else if c.ContentStore.Scheme == "memory" {
		return NewInMemoryContentState()
	} else {
		panic("Unsupported content store URL " + c.ContentStore.String())
	}
}

func (c *PlatformConfig) NewAuthState() AuthState {
	if c.AuthStore.Scheme == "memory" {
		return NewInMemoryAuthState()
	} else {
		panic("Unsupported content store URL " + c.AuthStore.String())
	}
}

func HackerNews(config *PlatformConfig) (*App, []Starter) {
	commandLog := config.NewCommandLog()
	contentState := config.NewContentState()
	content := NewContent(contentState)
	authState := config.NewAuthState()
	auth := NewAuth(authState)

	app := NewApp(commandLog)

	randomNumbers := rand.New(rand.NewSource(time.Now().UnixNano()))
	emailLogger := log.New(os.Stdout, "[mailer] ", log.LstdFlags)
	emailSender := NewFlakyEmailSender(&EmailLogger{emailLogger}, 0.95, randomNumbers)
	mailer := NewMailer(emailSender, emailLogger, app)

	previewLogger := log.New(os.Stdout, "[preview] ", log.LstdFlags)
	previewGenerator := NewPreviewGenerator(app, commandLog, previewLogger)
	starters := []Starter{previewGenerator, mailer}

	MustSetup(commandLog)
	MustSetup(auth)
	MustSetup(content)
	MustSetup(contentState)

	return app.Mount(auth).Mount(content), starters
}
