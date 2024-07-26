package main

import (
	"time"
)

type AuthState interface {
	SetUser(user *User) error
	FindUser(username string) (*User, error)
	FindSession(sessionID string) (*Session, error)
	SetSession(session *Session) error
}

type User struct {
	Username     string
	PasswordHash string
}

type Session struct {
	ID         string
	Username   string
	ActiveFrom time.Time
	ActiveTo   time.Time
}

func (s *Session) IsActive(now time.Time) bool {
	return now.After(s.ActiveFrom) && now.Before(s.ActiveTo)
}

type Auth struct {
	state AuthState
}

func NewAuth(state AuthState) *Auth {
	return &Auth{state: state}
}

func NewDefaultAuth() *Auth {
	return NewAuth(NewInMemoryAuthState())
}

func (self *Auth) HandleCommand(cmd Command) error {
	switch cmd := cmd.(type) {
	case *SignUpUser:
		return self.handleSignUpUser(cmd)
	case *LogInUser:
		return self.handleLogInUser(cmd)
	}
	return ErrCommandNotAccepted
}

func (self *Auth) HandleQuery(query Query) error {
	switch query := query.(type) {
	case *FindSession:
		return self.findSession(query)
	case *FindUserBySessionID:
		return self.FindUserBySessionID(query)
	default:
		return ErrQueryNotAccepted
	}
}
