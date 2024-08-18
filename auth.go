package main

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type AuthState interface {
	GetPolicy(policy *UsernamePolicy) error
	PutPolicy(policy *UsernamePolicy) error
	SetAdminUsers(users []string) error
	IsAdmin(username string) (bool, error)
	SetUser(user *User) error

	SetMagicDomains(domains []string) error
	GetMagicDomains() ([]string, error)

	FindUser(username string) (*User, error)
	FindUserByEmail(username string) (*User, error)
	FindUserByMagic(magic string) (*User, error)
	FindUserByPasswordResetToken(token string) (*User, error)
	FindSession(sessionID string) (*Session, error)

	SetSession(session *Session) error
}

type User struct {
	Username                 string
	PasswordHash             string
	VerifiedEmail            string
	Magic                    string
	PasswordResetToken       string
	PasswordResetRequestedAt time.Time
}

type Session struct {
	ID         string
	Username   string
	ActiveFrom time.Time
	ActiveTo   time.Time
}

type UsernamePolicy struct {
	MinLength int
	MaxLength int
	Blacklist []string
}

func (policy *UsernamePolicy) Allowed(username string) bool {
	if len(username) < policy.MinLength || len(username) > policy.MaxLength {
		return false
	}
	for _, blacklisted := range policy.Blacklist {
		if blacklisted == username {
			return false
		}
	}
	return true
}

func (s *Session) IsActive(now time.Time) bool {
	return now.After(s.ActiveFrom) && now.Before(s.ActiveTo)
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
	case *ChangeUsernamePolicy:
		return self.handleChangeUsernamePolicy(cmd)
	case *LinkVerifiedEmailToUser:
		return self.linkVerifiedEmailToUser(cmd)
	case *SetAdminUsers:
		return self.handleSetAdminUsers(cmd)
	case *RequestMagicLinkLogin:
		return self.handleRequestMagicLinkLogin(cmd)
	case *LogInWithMagic:
		return self.handleLogInUserWithMagic(cmd)
	case *SetMagicDomains:
		return self.handleSetMagicDomains(cmd)
	case *RequestPasswordReset:
		return self.handleRequestPasswordReset(cmd)
	case *ResetPassword:
		return self.handleResetPassword(cmd)
	}
	return ErrCommandNotAccepted
}

func (self *Auth) HandleQuery(query Query) error {
	switch query := query.(type) {
	case *FindSession:
		return self.findSession(query)
	case *FindUserBySessionID:
		return self.FindUserBySessionID(query)
	case *FindUserPasswordHash:
		return self.findUserPasswordHash(query)
	case *FindUserByEmail:
		return self.findUserByEmail(query)
	case *GetUserRoles:
		return self.getUserRoles(query)
	default:
		return ErrQueryNotAccepted
	}
}
