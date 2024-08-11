package main

import "fmt"

type UserRole string

var (
	UserRoleAdmin   UserRole = "admin"
	UserRoleVisitor UserRole = "visitor"
)

type GetUserRoles struct {
	Username string
	Roles    []UserRole
}

func (q *GetUserRoles) QueryName() string {
	return "GetUserRoles"
}

func NewGetUserRolesQuery(username string) *GetUserRoles {
	return &GetUserRoles{Username: username}
}

func (self *Auth) getUserRoles(q *GetUserRoles) error {
	user, err := self.state.FindUser(q.Username)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	q.Roles = []UserRole{}
	isAdmin, err := self.state.IsAdmin(q.Username)
	if err != nil {
		return fmt.Errorf("failed to check if user is admin: %w", err)
	}
	if isAdmin {
		q.Roles = append(q.Roles, UserRoleAdmin)
	}
	q.Roles = append(q.Roles, UserRoleVisitor)
	return nil
}
