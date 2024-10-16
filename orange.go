package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
)

type PlatformConfig struct {
	SkipErrorsDuringReplay  bool
	EmailSender             *url.URL
	ContentStore            *url.URL
	AuthStore               *url.URL
	CommandLog              *url.URL
	Notifier                *url.URL
	MagicLoginController    *url.URL
	PasswordResetController *url.URL
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
		SkipErrorsDuringReplay:  false,
		EmailSender:             parseURL("memory://", "EmailSender"),
		ContentStore:            parseURL("memory://", "ContentStore"),
		AuthStore:               parseURL("memory://", "AuthStore"),
		CommandLog:              parseURL("file:///commands.db", "CommandLog"),
		Notifier:                parseURL("service:///?baseUrl=http:%2f%2flocalhost:8081%2f", "Notifier"),
		MagicLoginController:    parseURL("service:///?baseUrl=http:%2f%2flocalhost:8081%2f", "MagicLoginController"),
		PasswordResetController: parseURL("service:///?baseUrl=http:%2f%2flocalhost:8081%2f", "PasswordResetController"),
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
		"CONTENT_STORE":             &config.ContentStore,
		"AUTH_STORE":                &config.AuthStore,
		"COMMAND_LOG":               &config.CommandLog,
		"EMAIL_SENDER":              &config.EmailSender,
		"NOTIFIER":                  &config.Notifier,
		"MAGIC_LOGIN_CONTROLLER":    &config.MagicLoginController,
		"PASSWORD_RESET_CONTROLLER": &config.PasswordResetController,
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
		from := c.EmailSender.Query().Get("from")
		if serverToken == "" {
			panic("Missing key in email sender URL")
		}
		if from == "" {
			panic("Missing from address in email sender URL")
		}
		logger := log.New(os.Stdout, "[postmark] ", log.LstdFlags)
		return NewPostmarkEmailSender(logger, c.EmailSender.Query())
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

func (c *PlatformConfig) NewNotifier(app *App) *Notifier {
	if c.MagicLoginController.Scheme == "service" {
		baseURL, err := url.Parse(c.Notifier.Query().Get("baseUrl"))
		if err != nil {
			panic("notifier: invalid base url" + err.Error())
		}
		return NewNotifier(
			app,
			app.Commands,
			log.New(os.Stdout, "[notifier] ", log.LstdFlags),
			baseURL,
		)
	}
	panic("Unsupported notifier URL " + c.MagicLoginController.String())
}

func (c *PlatformConfig) NewPasswordResetController(app *App) *PasswordResetController {
	if c.PasswordResetController.Scheme == "service" {
		return NewPasswordResetController(
			app,
			log.New(os.Stdout, "[password-reset] ", log.LstdFlags),
			c.PasswordResetController.Query().Get("baseUrl"),
		)
	}
	panic("Unsupported password reset controller URL " + c.PasswordResetController.String())
}

func HackerNews(config *PlatformConfig) (*App, []Starter) {
	commandLog := config.NewCommandLog()
	contentState := config.NewContentState()
	content := NewContent(contentState)
	authState := config.NewAuthState()
	auth := NewAuth(authState)

	app := NewApp(commandLog)

	magicLoginController := config.NewMagicLoginController(app)
	passwordResetController := config.NewPasswordResetController(app)

	emailLogger := log.New(os.Stdout, "[mailer] ", log.LstdFlags)
	emailSender := config.NewEmailSender()
	mailer := NewMailer(emailSender, emailLogger, app)

	notifier := config.NewNotifier(app)

	previewLogger := log.New(os.Stdout, "[preview] ", log.LstdFlags)
	previewGenerator := NewPreviewGenerator(app, commandLog, previewLogger)
	starters := []Starter{
		previewGenerator,
		mailer,
		magicLoginController,
		passwordResetController,
		notifier,
	}

	MustSetup(commandLog)
	MustSetup(auth)
	MustSetup(content)
	MustSetup(contentState)

	return app.Mount(auth).Mount(content), starters
}
