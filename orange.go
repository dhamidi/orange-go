package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
)

type PlatformConfig struct {
	SkipErrorsDuringReplay bool
	EmailSender            *url.URL
	ContentStore           *url.URL
	AuthStore              *url.URL
	CommandLog             *url.URL
	MagicLoginController   *url.URL
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
		EmailSender:            parseURL("memory://", "EmailSender"),
		ContentStore:           parseURL("memory://", "ContentStore"),
		AuthStore:              parseURL("memory://", "AuthStore"),
		CommandLog:             parseURL("file:///commands.db", "CommandLog"),
	}
}

func NewPlatformConfigForTest() *PlatformConfig {
	config := DefaultPlatformConfig()
	config.CommandLog = parseURL("memory://", "CommandLog")
	config.EmailSender = parseURL("memory://", "EmailSender")
	return config
}
func NewPlatformConfigFromEnv(getenv func(key string) string) *PlatformConfig {
	config := DefaultPlatformConfig()
	fields := map[string]**url.URL{
		"CONTENT_STORE": &config.ContentStore,
		"AUTH_STORE":    &config.AuthStore,
		"COMMAND_LOG":   &config.CommandLog,
		"EMAIL_SENDER":  &config.EmailSender,
	}
	for name, dest := range fields {
		newURL := getenv("ORANGE_" + name)
		if newURL == "" {
			continue
		}
		*dest = parseURL(newURL, name)
	}

	config.SkipErrorsDuringReplay = getenv("ORANGE_SKIP_ERRORS") == "true"

	return config
}

func (c *PlatformConfig) NewEmailSender() EmailSender {
	if c.EmailSender.Scheme == "memory" {
		return &EmailLogger{log.New(os.Stdout, "[mailer] ", log.LstdFlags)}
	}
	if c.EmailSender.Scheme == "https" && c.EmailSender.Host == "api.postmarkapp.com" {
		serverToken := c.EmailSender.Query().Get("key")
		logger := log.New(os.Stdout, "[postmark] ", log.LstdFlags)
		return NewPostmarkEmailSender(logger, serverToken)
	}

	panic("Unsupported email sender URL " + c.EmailSender.String())
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

func (c *PlatformConfig) NewMagicLoginController(app *App) *MagicLoginController {
	if c.MagicLoginController.Scheme == "service" {
		return NewMagicLoginController(
			app,
			app.Commands,
			log.New(os.Stdout, "[magic] ", log.LstdFlags),
			c.MagicLoginController.Query().Get("baseUrl"),
		)
	}
	panic("Unsupported magic login controller URL " + c.MagicLoginController.String())
}

func HackerNews(config *PlatformConfig) (*App, []Starter) {
	commandLog := config.NewCommandLog()
	contentState := config.NewContentState()
	content := NewContent(contentState)
	authState := config.NewAuthState()
	auth := NewAuth(authState)

	app := NewApp(commandLog)

	magicLoginController := config.NewMagicLoginController(app)
	emailLogger := log.New(os.Stdout, "[mailer] ", log.LstdFlags)
	emailSender := config.NewEmailSender()
	mailer := NewMailer(emailSender, emailLogger, app)

	previewLogger := log.New(os.Stdout, "[preview] ", log.LstdFlags)
	previewGenerator := NewPreviewGenerator(app, commandLog, previewLogger)
	starters := []Starter{previewGenerator, mailer, magicLoginController}

	MustSetup(commandLog)
	MustSetup(auth)
	MustSetup(content)
	MustSetup(contentState)

	return app.Mount(auth).Mount(content), starters
}
