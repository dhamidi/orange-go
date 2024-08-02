package main

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid_credentials")

type LogInUser struct {
	Username     string
	PasswordHash PasswordHash
	AttemptedAt  time.Time
	SessionID    string
}

func (cmd *LogInUser) CommandName() string {
	return "LogInUser"
}

func init() {
	DefaultCommandRegistry.Register("LogInUser", func() Command { return &LogInUser{} })
}

func (self *Auth) handleLogInUser(cmd *LogInUser) error {
	user, err := self.state.FindUser(cmd.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInvalidCredentials
	}

	if cmd.PasswordHash.String() != user.PasswordHash {
		return ErrInvalidCredentials
	}
	if err := self.state.SetSession(&Session{
		ID:         cmd.SessionID,
		Username:   cmd.Username,
		ActiveFrom: cmd.AttemptedAt,
		ActiveTo:   cmd.AttemptedAt.Add(time.Hour),
	}); err != nil {
		return fmt.Errorf("Failed to persist session: %w", err)
	}

	return nil
}
