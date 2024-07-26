package main

import (
	"fmt"
	"time"
)

type SignUpUser struct {
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

func (cmd *SignUpUser) CommandName() string {
	return "SignUpUser"
}

type LogInUser struct {
	Username     string
	PasswordHash string
	AttemptedAt  time.Time
	SessionID    string
}

func (cmd *LogInUser) CommandName() string {
	return "LogInUser"
}

func init() {
	DefaultSerializer.Register("SignUpUser", func() Command { return &SignUpUser{} })
	DefaultSerializer.Register("LogInUser", func() Command { return &LogInUser{} })
}

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

type InMemoryAuthState struct {
	Users    map[string]*User
	Sessions map[string]*Session
}

func NewInMemoryAuthState() *InMemoryAuthState {
	return &InMemoryAuthState{
		Users:    make(map[string]*User),
		Sessions: make(map[string]*Session),
	}
}

func (state *InMemoryAuthState) SetUser(user *User) error {
	state.Users[user.Username] = user
	return nil
}

func (state *InMemoryAuthState) SetSession(session *Session) error {
	state.Sessions[session.ID] = session
	return nil
}

func (state *InMemoryAuthState) FindSession(sessionID string) (*Session, error) {
	return state.Sessions[sessionID], nil
}

func (state *InMemoryAuthState) FindUser(username string) (*User, error) {
	return state.Users[username], nil
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

func (self *Auth) handleSignUpUser(cmd *SignUpUser) error {
	user, err := self.state.FindUser(cmd.Username)
	if err != nil {
		return err
	}
	if user != nil {
		return fmt.Errorf("user already exists")
	}
	return self.state.SetUser(&User{
		Username:     cmd.Username,
		PasswordHash: cmd.PasswordHash,
	})
}

func (self *Auth) handleLogInUser(cmd *LogInUser) error {
	user, err := self.state.FindUser(cmd.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("invalid_credentials")
	}
	passwordMatches, err := ComparedPasswordHashes(cmd.PasswordHash, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("Failed to compare password to hash: %w", err)
	}
	if !passwordMatches {
		return fmt.Errorf("invalid_credentials")
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

type FindSession struct {
	ID      string
	Session *Session
}

func (q *FindSession) QueryName() string {
	return "FindSession"
}

func (self *Auth) HandleQuery(query Query) error {
	switch query := query.(type) {
	case *FindSession:
		session, err := self.state.FindSession(query.ID)
		if err != nil {
			return err
		}
		query.Session = session
	default:
		return ErrQueryNotAccepted
	}
	return nil
}

func NewFindSessionQuery(id string) *FindSession {
	return &FindSession{ID: id}
}
