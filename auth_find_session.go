package main

type FindSession struct {
	ID      string
	Session *Session
}

func (q *FindSession) QueryName() string {
	return "FindSession"
}

func NewFindSessionQuery(id string) *FindSession {
	return &FindSession{ID: id}
}

func (self *Auth) findSession(query *FindSession) error {
	session, err := self.state.FindSession(query.ID)
	if err != nil {
		return err
	}
	query.Session = session
	return nil
}
