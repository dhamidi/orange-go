package main

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Shell struct {
	CurrentTime func() time.Time
	NewID       func() string
	App         *App
}

func NewDefaultShell(app *App) *Shell {
	return &Shell{
		CurrentTime: time.Now,
		NewID:       uuid.NewString,
		App:         app,
	}
}

type Parameters interface {
	Get(key string) string
}

func GetAllValues(p Parameters, key string) []string {
	values := []string{}
	for i := 0; true; i++ {
		value := p.Get(fmt.Sprintf("%s[%d]", key, i))
		if value == "" {
			return values
		}
		values = append(values, value)
	}
	return values
}

func (s *Shell) Signup(params Parameters) error {
	username, password := params.Get("username"), params.Get("password")

	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("signup: failed to hash password: %w", err)
	}
	signup := &SignUpUser{
		Username:     username,
		PasswordHash: *passwordHash,
		CreatedAt:    s.CurrentTime(),
	}
	if err := s.App.HandleCommand(signup); err != nil {
		return fmt.Errorf("signup: %w", err)
	}
	return nil
}

func (s *Shell) Login(params Parameters) (string, error) {
	username, password := params.Get("username"), params.Get("password")
	q := NewFindUserPasswordHash(username, password)
	if err := s.App.HandleQuery(q); err != nil {
		return "", fmt.Errorf("login: failed to hash password: %w\n", err)
	}
	if q.PasswordHash == nil {
		return "", ErrInvalidCredentials
	}
	login := &LogInUser{
		Username:     username,
		PasswordHash: *q.PasswordHash,
		AttemptedAt:  s.CurrentTime(),
		SessionID:    s.NewID(),
	}
	if err := s.App.HandleCommand(login); err != nil {
		return "", fmt.Errorf("failed to log in user: %w\n", err)

	}
	return login.SessionID, nil
}

func (s *Shell) RequestMagicLinkLogin(params Parameters) (string, error) {
	email := params.Get("email")
	magic := s.NewID()
	request := &RequestMagicLinkLogin{
		Email:       email,
		Magic:       magic,
		RequestedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(request); err != nil {
		return "", fmt.Errorf("request-magic-link-login: %w\n", err)
	}
	return request.Magic, nil
}

func (s *Shell) LoginWithMagicLink(params Parameters) (string, error) {
	magic := params.Get("magic")
	sessionID := s.NewID()
	login := &LogInWithMagic{
		SessionID:   sessionID,
		Magic:       magic,
		AttemptedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(login); err != nil {
		return "", fmt.Errorf("login-with-magic-link: %w\n", err)
	}
	return sessionID, nil
}

func (s *Shell) LinkVerifiedEmailToUser(params Parameters) error {
	username := params.Get("username")
	email := params.Get("email")
	link := &LinkVerifiedEmailToUser{
		Username: username,
		Email:    email,
		LinkedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(link); err != nil {
		return fmt.Errorf("link-verified-email-to-user: %w\n", err)
	}
	return nil
}

func (s *Shell) SetAdminUsers(params Parameters) error {
	usernames := GetAllValues(params, "username")
	setAdmin := &SetAdminUsers{Users: usernames}
	if err := s.App.HandleCommand(setAdmin); err != nil {
		return fmt.Errorf("set-admin-users: %w\n", err)
	}
	return nil
}

func (s *Shell) SetMagicDomains(params Parameters) error {
	domains := GetAllValues(params, "domain")
	setMagicDomains := &SetMagicDomains{Domains: domains}
	if err := s.App.HandleCommand(setMagicDomains); err != nil {
		return fmt.Errorf("set-magic-domains: %w\n", err)
	}
	return nil
}

func (s *Shell) GetUserRoles(params Parameters) ([]UserRole, error) {
	username := params.Get("username")
	q := NewGetUserRolesQuery(username)
	if err := s.App.HandleQuery(q); err != nil {
		return nil, fmt.Errorf("get-user-roles: failed to get user roles: %w\n", err)
	}
	return q.Roles, nil
}

func (s *Shell) FindSession(params Parameters) (*Session, error) {
	sessionID := params.Get("sessionID")
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return nil, fmt.Errorf("find-session: failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return nil, ErrSessionNotFound
	}
	return q.Session, nil
}

func (s *Shell) UnskipCommands(params Parameters) error {
	ids := GetAllValues(params, "id")
	reviser, ok := s.App.Commands.(CommandReviser)
	if !ok {
		return fmt.Errorf("Command log type %T does not support revision", s.App.Commands)
	}
	intIDs := []int{}
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			return fmt.Errorf("failed to convert id[%d] %q to int: %w", i, id, err)
		}
		intIDs = append(intIDs, i)
	}
	reviser.ReviseCommands(intIDs, func(id int) Command { return nil })
	return nil
}

func (s *Shell) SkipCommands(params Parameters) error {
	ids := GetAllValues(params, "id")
	reviser, ok := s.App.Commands.(CommandReviser)
	if !ok {
		return fmt.Errorf("Command log type %T does not support revision", s.App.Commands)
	}
	intIDs := []int{}
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			return fmt.Errorf("failed to convert id[%d] %q to int: %w", i, id, err)
		}
		intIDs = append(intIDs, i)
	}
	skip := new(SkipCommand)
	reviser.ReviseCommands(intIDs, func(id int) Command { return skip })
	return nil
}

func (s *Shell) List(params Parameters, out io.Writer) error {
	after := 0
	if n := params.Get("after"); n != "" {
		i, err := strconv.Atoi(n)
		if err == nil {
			after = i
		} else {
			return fmt.Errorf("list: %w", err)
		}
	}
	log := s.App.Commands
	commands, err := log.After(after)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}

	formatFieldValue := func(v reflect.Value) string {
		switch v.Kind() {
		case reflect.Struct:
			if t, ok := v.Interface().(fmt.Stringer); ok {
				return fmt.Sprintf("%q", t)
			}
			if v.CanAddr() {
				if t, ok := v.Addr().Interface().(fmt.Stringer); ok {
					return fmt.Sprintf("%q", t)
				}
			}
			return fmt.Sprintf("%#v", v.Interface())
		case reflect.Pointer:
			if v.IsZero() {
				return "<nil>"
			} else {
				return fmt.Sprintf("%q", v.Elem())
			}
		default:
			return fmt.Sprintf("%q", v.Interface())
		}
	}
	formatPayload := func(command Command) string {
		c := reflect.ValueOf(command).Elem() // &PostLink{...} -> PostLink{...}
		t := c.Type()                        // PostLink{...} -> PostLink
		buffer := bytes.NewBufferString("")
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fmt.Fprintf(buffer, " %s=%s", field.Name, formatFieldValue(c.Field(i)))
		}
		return buffer.String()
	}

	for command := range commands {
		fmt.Fprintf(out, "%3d %20s %s\n", command.ID, command.Message.CommandName(), formatPayload(command.Message))
	}
	return nil
}

func (s *Shell) Upvote(params Parameters) error {
	sessionID := params.Get("sessionID")
	itemID := params.Get("itemID")
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return ErrSessionNotFound
	}
	submit := &UpvoteSubmission{
		ItemID:  itemID,
		Voter:   q.Session.Username,
		VotedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(submit); err != nil {
		return fmt.Errorf("upvote: %w", err)
	}
	return nil
}

func (s *Shell) Submit(params Parameters) error {
	sessionID := params.Get("sessionID")
	title := params.Get("title")
	url := params.Get("url")
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return ErrSessionNotFound
	}
	submit := &PostLink{
		ItemID:      s.NewID(),
		Submitter:   q.Session.Username,
		Title:       title,
		Url:         url,
		SubmittedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(submit); err != nil {
		return fmt.Errorf("failed to submit content: %w\n", err)
	}
	return nil
}

func (s *Shell) RequestPasswordReset(params Parameters) (string, error) {
	email := params.Get("email")
	token := s.NewID()

	q := NewFindUserByEmail(email)
	if err := s.App.HandleQuery(q); err != nil || q.User == nil {
		return "", ErrUserNotFound
	}

	reset := &RequestPasswordReset{
		Username:    q.User.Username,
		Email:       email,
		Token:       token,
		RequestedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(reset); err != nil {
		return "", fmt.Errorf("failed to request password reset: %w\n", err)
	}
	return token, nil
}

func (s *Shell) ResetPassword(params Parameters) error {
	token := params.Get("token")
	newPassword, err := HashPassword(params.Get("password"))
	if err != nil {
		return fmt.Errorf("reset-password: failed to hash password: %w", err)
	}
	reset := &ResetPassword{
		Token:       token,
		NewPassword: *newPassword,
		AttemptedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(reset); err != nil {
		return fmt.Errorf("failed to reset password: %w\n", err)
	}
	return nil
}

func (s *Shell) Comment(params Parameters) error {
	sessionID := params.Get("sessionID")
	itemID := params.Get("itemID")
	text := params.Get("text")

	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return ErrSessionNotFound
	}
	commentOn := &PostComment{
		ParentID: NewTreeID(itemID),
		Author:   q.Session.Username,
		Content:  text,
		PostedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(commentOn); err != nil {
		return fmt.Errorf("failed to post comment: %w\n", err)
	}
	return nil
}
