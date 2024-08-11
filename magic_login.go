package main

import (
	"log"
	"time"
)

type MagicLoginController struct {
	Logger   *log.Logger
	App      *App
	Commands CommandLog
}

func NewMagicLoginController(app *App, commands CommandLog, logger *log.Logger) *MagicLoginController {
	return &MagicLoginController{
		Logger:   logger,
		App:      app,
		Commands: commands,
	}
}

func (m *MagicLoginController) HandleCommand(command Command) error {
	switch c := command.(type) {
	case *RequestMagicLinkLogin:
		return m.handleMagicLogin(c)
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

	// TODO: weird place
	for c := range commands {
		m.HandleCommand(c.Message)
	}

	go m.loop(stop)
	return func() { close(stop) }
}

func (m *MagicLoginController) handleMagicLogin(c *RequestMagicLinkLogin) error {
	panic("implement me")
}

func (p *MagicLoginController) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	tick := ticker.C
	for {
		select {
		case <-stop:
			return
		case <-tick:
			// p.fetchPreviews()
			panic("implement me")
		}
	}
}
