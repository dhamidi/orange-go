package main

import (
	"iter"
	"log"
	"time"
)

type RequestState struct {
	Magic       string
	IsCompleted bool
	RetriesLeft uint
}

type MagicLoginController struct {
	Logger          *log.Logger
	App             *App
	Commands        CommandLog
	MailClient      MailAgent
	NumberOfRetries uint
	PendingRequests map[string]RequestState
}

func NewMagicLoginController(app *App, commands CommandLog, logger *log.Logger, retries uint) *MagicLoginController {
	return &MagicLoginController{
		Logger:          logger,
		App:             app,
		Commands:        commands,
		NumberOfRetries: retries,
		PendingRequests: make(map[string]RequestState),
	}
}

func (m *MagicLoginController) HandleCommand(command Command) error {
	switch c := command.(type) {
	case *RequestMagicLinkLogin:
		return m.addPendingRequest(c)
	default:
		return ErrCommandNotAccepted
	}
}

func (m *MagicLoginController) Start() func() {
	stop := make(chan struct{})

	commands, err := m.Commands.After(0)
	if err != nil {
		m.Logger.Printf("failed to fetch commands: %v", err)
		return func() {}
	}

	then := time.Now().Add(time.Duration(-1) * time.Hour)
	// Merge latest requests with pending request ??
	latestRequests := getRequestsUntil(commands, then)

	for c := range latestRequests {
		m.HandleCommand(c.Message)
	}

	go m.loop(stop)
	return func() { close(stop) }
}

func (p *MagicLoginController) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	tick := ticker.C
	for {
		select {
		case <-stop:
			return
		case <-tick:
			p.handleMagicLogin()
		}
	}
}

func (m *MagicLoginController) addPendingRequest(c *RequestMagicLinkLogin) error {
	m.PendingRequests[c.Email] = RequestState{
		Magic:       c.Magic,
		IsCompleted: false,
		RetriesLeft: m.NumberOfRetries,
	}
	return nil
}

func (m *MagicLoginController) handleMagicLogin() error {
	// for email, requestState := range m.PendingRequests {

	// }
	return nil
}

func getRequestsUntil(commands iter.Seq[*PersistedCommand], until time.Time) iter.Seq[*PersistedCommand] {
	panic("implement me")
}
