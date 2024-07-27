package main

import "fmt"

type FindUserPasswordHash struct {
	Username     string
	Password     string
	PasswordHash *PasswordHash
}

func (q *FindUserPasswordHash) QueryName() string {
	return "FindUserPasswordHash"
}

func NewFindUserPasswordHash(username, password string) *FindUserPasswordHash {
	return &FindUserPasswordHash{Username: username, Password: password}
}

func (self *Auth) findUserPasswordHash(q *FindUserPasswordHash) error {
	user, err := self.state.FindUser(q.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	hash := new(PasswordHash)
	if err := hash.UnmarshalText([]byte(user.PasswordHash)); err != nil {
		return fmt.Errorf("failed to unmarshal password hash: %s", err)
	}

	if err := hash.Check(q.Password); err != nil {
		return err
	}

	q.PasswordHash = hash
	return nil
}
