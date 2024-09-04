package main

type FindUserBySessionID struct {
	SessionID string
	User      *User
}

func (q *FindUserBySessionID) QueryName() string {
	return "FindUserBySessionID"
}

func (q *FindUserBySessionID) Result() any { return q.User }

func NewFindUserBySessionIDQuery(sessionID string) *FindUserBySessionID {
	return &FindUserBySessionID{SessionID: sessionID}
}

func (self *Auth) FindUserBySessionID(query *FindUserBySessionID) error {
	session, err := self.state.FindSession(query.SessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}
	user, err := self.state.FindUser(session.Username)
	if err != nil {
		return err
	}
	query.User = user
	return nil
}
