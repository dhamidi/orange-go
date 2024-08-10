package main

type InMemoryAuthState struct {
	UsernamePolicy *UsernamePolicy
	Users          map[string]*User
	Sessions       map[string]*Session
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
