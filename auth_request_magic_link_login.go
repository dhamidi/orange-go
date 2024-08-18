package main

import (
	"errors"
	"strings"
	"time"
)

type RequestMagicLinkLogin struct {
	Email       string
	Magic       string
	RequestedAt time.Time
}

func (cmd *RequestMagicLinkLogin) DomainIsMagic(magicDomains []string) bool {
	for _, domain := range magicDomains {
		if strings.Contains(cmd.Email, "@"+domain) {
			return true
		}
	}
	return false
}

func (cmd *RequestMagicLinkLogin) UsernameFromEmail() string {
	return strings.Split(cmd.Email, "@")[0]
}

func (cmd *RequestMagicLinkLogin) CommandName() string {
	return "RequestMagicLinkLogin"
}

func init() {
	DefaultCommandRegistry.Register("RequestMagicLinkLogin", func() Command { return new(RequestMagicLinkLogin) })
}

func (self *Auth) handleRequestMagicLinkLogin(cmd *RequestMagicLinkLogin) error {
	magicDomains, err := self.state.GetMagicDomains()
	if err != nil {
		return err
	}
	user, err := self.state.FindUserByEmail(cmd.Email)
	if errors.Is(err, ErrUserNotFound) && cmd.DomainIsMagic(magicDomains) {
		return self.signUpUserByMagic(cmd)
	} else if err != nil {
		return err
	}

	user.Magic = cmd.Magic
	return self.state.SetUser(user)
}

func (self *Auth) signUpUserByMagic(cmd *RequestMagicLinkLogin) error {
	user := &User{
		Username:      cmd.UsernameFromEmail(),
		Magic:         cmd.Magic,
		VerifiedEmail: cmd.Email,
	}
	return self.state.SetUser(user)
}
