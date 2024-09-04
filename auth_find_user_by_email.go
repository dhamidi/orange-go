package main

type FindUserByEmail struct {
	Email string
	User  *User
}

func (q *FindUserByEmail) QueryName() string { return "FindUserByEmail" }
func (q *FindUserByEmail) Result() any       { return q.User }

func NewFindUserByEmail(email string) *FindUserByEmail {
	return &FindUserByEmail{Email: email}
}

func (self *Auth) findUserByEmail(q *FindUserByEmail) error {
	user, err := self.state.FindUserByEmail(q.Email)
	if err != nil {
		return err
	}
	q.User = user
	return nil
}
