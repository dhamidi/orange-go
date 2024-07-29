package main

import (
	"fmt"
	"time"
)

var ErrUserExists = fmt.Errorf("user already exists")

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
	return self.state.SetUser(&User{
		Username:     cmd.Username,
		PasswordHash: cmd.PasswordHash,
	})
}
