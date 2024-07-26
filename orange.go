package main

import (
	"fmt"
	"log"
	"net/url"
)

type PlatformConfig struct {
	ContentStore *url.URL
	AuthStore    *url.URL
	CommandLog   *url.URL
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
		ContentStore: parseURL("memory://", "ContentStore"),
		AuthStore:    parseURL("memory://", "AuthStore"),
		CommandLog:   parseURL("file:///commands.db", "CommandLog"),
	}
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
	return config
}

func (c *PlatformConfig) NewCommandLog() CommandLog {
	if c.CommandLog.Scheme == "file" {
		return NewFileCommandLog(toFilePath(c.CommandLog), DefaultSerializer)
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

	previewGenerator := NewPreviewGenerator(app, commandLog, log.Default())
	starters := []Starter{previewGenerator}

	MustSetup(commandLog)
	MustSetup(auth)
	MustSetup(content)
	MustSetup(contentState)

	return app.Mount(auth).Mount(content), starters
}
