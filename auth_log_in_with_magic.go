package main

import (
	"fmt"
	"time"
)

type LogInWithMagic struct {
	SessionID   string
	Magic       string
	AttemptedAt time.Time
}

func (cmd *LogInWithMagic) CommandName() string {
	return "LogInWithMagic"
}

func init() {
	DefaultCommandRegistry.Register("LogInWithMagic", func() Command { return new(LogInWithMagic) })
}

func (self *Auth) handleLogInUserWithMagic(cmd *LogInWithMagic) error {
	user, err := self.state.FindUserByMagic(cmd.Magic)
	if err != nil {
		return err
	}

	err = self.state.SetSession(&Session{
		ID:         cmd.SessionID,
		Username:   user.Username,
		ActiveFrom: cmd.AttemptedAt,
		ActiveTo:   cmd.AttemptedAt.Add(time.Hour),
	})

	// erase the magic link after it's been used
	user.Magic = ""

	if err != nil {
		return fmt.Errorf("Failed to persist session: %w", err)
	}

	return nil
}
