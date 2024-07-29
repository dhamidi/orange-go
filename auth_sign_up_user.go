package main

import (
	"fmt"
	"time"
)

var ErrUserExists = fmt.Errorf("user already exists")
var ErrUsernameNotAllowed = fmt.Errorf("username not allowed")

type SignUpUser struct {
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

func (cmd *SignUpUser) CommandName() string {
	return "SignUpUser"
}

func init() {
	DefaultSerializer.Register("SignUpUser", func() Command { return &SignUpUser{} })
}

func (self *Auth) handleSignUpUser(cmd *SignUpUser) error {
	user, err := self.state.FindUser(cmd.Username)
	if err != nil {
		return err
	}

	if user != nil {
		return ErrUserExists
	}

	policy := new(UsernamePolicy)
	if err := self.state.GetPolicy(policy); err != nil {
		return fmt.Errorf("failed to get username policy: %w", err)
	}

	if !policy.Allowed(cmd.Username) {
		return ErrUsernameNotAllowed
	}

	return self.state.SetUser(&User{
		Username:     cmd.Username,
		PasswordHash: cmd.PasswordHash,
	})
}
