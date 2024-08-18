package main

import (
	"maps"
	"slices"
)

var _ AuthState = &InMemoryAuthState{}

type InMemoryAuthState struct {
	UsernamePolicy *UsernamePolicy
	Users          map[string]*User
	AdminUsers     map[string]bool
	MagicDomains   map[string]bool
	Sessions       map[string]*Session
}

func (state *InMemoryAuthState) GetMagicDomains() ([]string, error) {
	return slices.Collect(maps.Keys(state.MagicDomains)), nil
}

func (state *InMemoryAuthState) SetMagicDomains(domains []string) error {
	state.MagicDomains = make(map[string]bool)
	for _, domain := range domains {
		state.MagicDomains[domain] = true
	}
	return nil
}

func NewInMemoryAuthState() *InMemoryAuthState {
	return &InMemoryAuthState{
		AdminUsers: make(map[string]bool),
		Users:      make(map[string]*User),
		Sessions:   make(map[string]*Session),
	}
}

func (state *InMemoryAuthState) SetAdminUsers(users []string) error {
	state.AdminUsers = make(map[string]bool)
	for _, user := range users {
		state.AdminUsers[user] = true
	}
	return nil
}

func (state *InMemoryAuthState) IsAdmin(username string) (bool, error) {
	return state.AdminUsers[username], nil
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

func (state *InMemoryAuthState) PutPolicy(policy *UsernamePolicy) error {
	state.UsernamePolicy = policy
	return nil
}

func (state *InMemoryAuthState) GetPolicy(policy *UsernamePolicy) error {
	if state.UsernamePolicy == nil {
		*policy = UsernamePolicy{
			MinLength: 0,
			MaxLength: 20,
		}
		return nil
	}
	*policy = *state.UsernamePolicy
	return nil
}

func (state *InMemoryAuthState) FindUserByEmail(email string) (*User, error) {
	for _, user := range state.Users {
		if user.VerifiedEmail == email {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (state *InMemoryAuthState) FindUserByMagic(magic string) (*User, error) {
	for _, user := range state.Users {
		if user.Magic == magic {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (state *InMemoryAuthState) FindUserByPasswordResetToken(token string) (*User, error) {
	for _, user := range state.Users {
		if user.PasswordResetToken == token {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}
