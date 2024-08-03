package main

import (
	"fmt"
	"time"
)

type LinkVerifiedEmailToUser struct {
	Email    string
	Username string
	LinkedAt time.Time
}

func (cmd *LinkVerifiedEmailToUser) CommandName() string {
	return "LinkVerifiedEmailToUser"
}

func init() {
	DefaultCommandRegistry.Register("LinkVerifiedEmailToUser", func() Command { return &LinkVerifiedEmailToUser{} })
}

func (self *Auth) linkVerifiedEmailToUser(cmd *LinkVerifiedEmailToUser) error {
	user, err := self.state.FindUser(cmd.Username)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	user.VerifiedEmail = cmd.Email
	return self.state.SetUser(user)
}
