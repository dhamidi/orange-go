package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Shell struct {
	CurrentTime func() time.Time
	NewID       func() string
	Output      io.Writer
	App         *App
}

func NewDefaultShell(app *App) *Shell {
	return &Shell{
		CurrentTime: time.Now,
		NewID:       uuid.NewString,
		Output:      os.Stdout,
		App:         app,
	}
}

type Parameters interface {
	Get(key string) string
}

func (s *Shell) Info(format string, args ...interface{}) {
	fmt.Fprintf(s.Output, format, args...)
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
