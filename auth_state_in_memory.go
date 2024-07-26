package main

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
