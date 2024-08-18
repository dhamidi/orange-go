package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PasswordResetController struct {
	BaseUrl         string
	Logger          *log.Logger
	App             *App
	PendingRequests map[string]string
	Version         int
}

func NewPasswordResetController(app *App, logger *log.Logger, baseUrl string) *PasswordResetController {
	return &PasswordResetController{
		BaseUrl:         baseUrl,
		Logger:          logger,
		App:             app,
		PendingRequests: make(map[string]string),
		Version:         0,
	}
}

func (p *PasswordResetController) HandleCommand(command Command, from time.Time) error {
	switch c := command.(type) {
	case *RequestPasswordReset:
		return p.addPendingRequest(c, from)
	case *ResetPassword:
		return p.removePendingRequest(c.Token)
	case *QueueEmail:
		return p.handleQueueEmail(c)
	default:
		return ErrCommandNotAccepted
	}
}

func (p *PasswordResetController) Start() func() {
	stop := make(chan struct{})
	p.catchUp()
	go p.loop(stop)
	return func() { close(stop) }
}

func (p *PasswordResetController) catchUp() {
	commands, err := p.App.Commands.After(p.Version)
	if err != nil {
		p.Logger.Printf("failed to fetch commands: %v", err)
		return
	}

	from := time.Now().Add(time.Duration(-1) * time.Hour)
	for c := range commands {
		p.HandleCommand(c.Message, from)
		p.Version = c.ID
	}
}

func (p *PasswordResetController) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	tick := ticker.C
	for {
		select {
		case <-stop:
			return
		case <-tick:
			p.catchUp()
			p.handlePasswordReset()
		}
	}
}

func (p *PasswordResetController) addPendingRequest(c *RequestPasswordReset, from time.Time) error {
	if c.RequestedAt.After(from) {
		p.PendingRequests[c.Email] = c.Token
	}
	return nil
}

func (p *PasswordResetController) removePendingRequest(token string) error {
	for email, t := range p.PendingRequests {
		if t == token {
			delete(p.PendingRequests, email)
			return nil
		}
	}
	return nil
}

func (p *PasswordResetController) handleQueueEmail(c *QueueEmail) error {
	if c.TemplateName != "password-reset" {
		return nil
	}
	actionUrl, ok := c.TemplateData["action_url"].(string)
	if !ok {
		return nil
	}
	pathSegments := strings.Split(actionUrl, "/")
	token := pathSegments[len(pathSegments)-1]
	p.removePendingRequest(token)
	return nil
}

func (p *PasswordResetController) handlePasswordReset() error {
	toBeRemoved := []string{}

	for email, token := range p.PendingRequests {
		query := NewFindUserByEmail(email)
		if err := p.App.HandleQuery(query); err == nil && query.User != nil {
			queueEmail := QueueEmail{
				InternalID:   uuid.NewString(),
				Recipients:   email,
				Subject:      "",
				TemplateName: "password-reset",
				TemplateData: map[string]any{
					"name":       query.User.Username,
					"action_url": fmt.Sprintf("%s/reset-password/%s", p.BaseUrl, token),
				},
			}
			p.App.HandleCommand(&queueEmail)
		}
		toBeRemoved = append(toBeRemoved, email)
	}

	for _, email := range toBeRemoved {
		delete(p.PendingRequests, email)
	}

	return nil
}
