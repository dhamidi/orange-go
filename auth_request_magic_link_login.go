package main

import (
	"time"
)

type RequestMagicLinkLogin struct {
	Email       string
	Magic       string
	RequestedAt time.Time
}

func (cmd *RequestMagicLinkLogin) CommandName() string {
	return "RequestMagicLinkLogin"
}

func init() {
	DefaultCommandRegistry.Register("RequestMagicLinkLogin", func() Command { return new(RequestMagicLinkLogin) })
}

func (self *Auth) handleRequestMagicLinkLogin(cmd *RequestMagicLinkLogin) error {
	user, err := self.state.FindUserByEmail(cmd.Email)

	if err != nil {
		return err
	}

	user.Magic = cmd.Magic
	return self.state.SetUser(user)
}
