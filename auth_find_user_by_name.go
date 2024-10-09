package main

type FindUserByName struct {
	Name string
	User *User
}

func (q *FindUserByName) QueryName() string { return "FindUserByName" }
func (q *FindUserByName) Result() any       { return q.User }

func NewFindUserByName(email string) *FindUserByName {
	return &FindUserByName{Name: email}
}

func (self *Auth) findUserByName(q *FindUserByName) error {
	user, err := self.state.FindUser(q.Name)
	if err != nil {
		return err
	}
	q.User = user
	return nil
}
