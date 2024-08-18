package main

import "time"

type RequestPasswordReset struct {
	Username    string
	Email       string
	Token       string
	RequestedAt time.Time
}

func (cmd *RequestPasswordReset) CommandName() string {
	return "RequestPasswordReset"
}

func init() {
	DefaultCommandRegistry.Register("RequestPasswordReset", func() Command { return new(RequestPasswordReset) })
}

func (self *Auth) handleRequestPasswordReset(cmd *RequestPasswordReset) error {
	user, err := self.state.FindUser(cmd.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.VerifiedEmail != cmd.Email {
		return ErrUserNotFound
	}
	user.PasswordResetToken = cmd.Token
	user.PasswordResetRequestedAt = cmd.RequestedAt
	return self.state.SetUser(user)
}
