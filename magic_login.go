package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type MagicLoginController struct {
	BaseUrl         string
	Logger          *log.Logger
	App             *App
	Commands        CommandLog
	PendingRequests map[string]string
	Version         int
}

func NewMagicLoginController(app *App, commands CommandLog, logger *log.Logger, baseUrl string) *MagicLoginController {
	return &MagicLoginController{
		BaseUrl:         baseUrl,
		Logger:          logger,
		App:             app,
		Commands:        commands,
		PendingRequests: make(map[string]string),
		Version:         0,
	}
}

func (m *MagicLoginController) HandleCommand(command Command, from time.Time) error {
	switch c := command.(type) {
	case *RequestMagicLinkLogin:
		return m.addPendingRequest(c, from)
	default:
		return ErrCommandNotAccepted
	}
}

func (m *MagicLoginController) Start() func() {
	stop := make(chan struct{})
	m.catchUp()
	go m.loop(stop)
	return func() { close(stop) }
}

func (m *MagicLoginController) catchUp() {
	commands, err := m.Commands.After(m.Version)
	if err != nil {
		m.Logger.Printf("failed to fetch commands: %v", err)
		return
	}

	from := time.Now().Add(time.Duration(-1) * time.Hour)
	for c := range commands {
		m.HandleCommand(c.Message, from)
		m.Version = c.ID
	}
}

func (m *MagicLoginController) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	tick := ticker.C
	for {
		select {
		case <-stop:
			return
		case <-tick:
			m.catchUp()
			m.handleMagicLogin()
		}
	}
}

func (m *MagicLoginController) addPendingRequest(c *RequestMagicLinkLogin, from time.Time) error {
	if c.RequestedAt.After(from) {
		m.PendingRequests[c.Email] = c.Magic
	}
	return nil
}

func (m *MagicLoginController) handleMagicLogin() error {
	toBeRemoved := []string{}

	for email, magic := range m.PendingRequests {
		query := NewFindUserByEmail(email)
		if err := m.App.HandleQuery(query); err == nil && query.User != nil {
			queueEmail := QueueEmail{
				InternalID:   uuid.NewString(),
				Recipients:   email,
				Subject:      "",
				TemplateName: "magic-login",
				TemplateData: map[string]any{
					"name":       query.User.Username,
					"action_url": fmt.Sprintf("%s/login/%s", m.BaseUrl, magic),
				},
			}
			m.App.HandleCommand(&queueEmail)
		}
		toBeRemoved = append(toBeRemoved, email)
	}

	for _, email := range toBeRemoved {
		delete(m.PendingRequests, email)
	}

	return nil
}
