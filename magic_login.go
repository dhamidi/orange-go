package main

import "log"

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
	case *AuthMagicLinkLogin:
		return m.handleMagicLogin(c)
	default:
		return ErrCommandNotAccepted
	}
}

func (m *MagicLoginController) Start() func() {
	panic("implement me")
}

func (m *MagicLoginController) handleMagicLogin(c *AuthMagicLinkLogin) error {
	panic("implement me")
}
