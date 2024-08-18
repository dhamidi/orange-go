package main

import (
	"errors"
	"time"
)

var ErrPasswordResetExpired = errors.New("password reset expired")

type ResetPassword struct {
	Token       string
	NewPassword PasswordHash
	AttemptedAt time.Time
}

func (cmd *ResetPassword) CommandName() string {
	return "ResetPassword"
}

func init() {
	DefaultCommandRegistry.Register("ResetPassword", func() Command { return new(ResetPassword) })
}

func (self *Auth) handleResetPassword(cmd *ResetPassword) error {
	user, err := self.state.FindUserByPasswordResetToken(cmd.Token)
	if err != nil {
		return err
	}

	if cmd.AttemptedAt.Sub(user.PasswordResetRequestedAt) >= 30*time.Minute {
		return ErrPasswordResetExpired
	}

	user.PasswordHash = cmd.NewPassword.String()
	user.PasswordResetToken = ""
	return self.state.SetUser(user)
}
