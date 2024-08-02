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

func (s *Shell) FindSession(sessionID string) {
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		s.Info("failed to find session %q: %s\n", sessionID, err)
		return
	}
	s.Info("%+v\n", q.Session)
}

func (s *Shell) List(params []string) {
	after := 0
	if len(params) > 0 {
		i, err := strconv.Atoi(params[0])
		if err == nil {
			after = i
		}
	}
	log := s.App.Commands
	commands, err := log.After(after)
	if err != nil {
		s.Info("failed to list commands: %s\n", err)
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
		s.Info("%3d %20s %s\n", command.ID, command.Message.CommandName(), formatPayload(command.Message))
	}
}

func (s *Shell) Upvote(params []string) {
	if len(params) < 2 {
		s.Info("usage: upvote <session-id> <item-id>\n")
		return
	}
	sessionID, itemID := params[0], params[1]
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		s.Info("failed to find session %q: %s\n", sessionID, err)
		return
	}
	if q.Session == nil {
		s.Info("You are not logged in\n")
		return
	}
	submit := &UpvoteSubmission{
		ItemID:  itemID,
		Voter:   q.Session.Username,
		VotedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(submit); err != nil {
		s.Info("failed to upvote: %s\n", err)
	}
}

func (s *Shell) Submit(params []string) {
	if len(params) < 3 {
		s.Info("usage: submit <session-id> <title> <url>\n")
		return
	}
	sessionID, title, url := params[0], params[1], params[2]
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		s.Info("failed to find session %q: %s\n", sessionID, err)
		return
	}
	if q.Session == nil {
		s.Info("You are not logged in\n")
		return
	}
	submit := &PostLink{
		ItemID:      s.NewID(),
		Submitter:   q.Session.Username,
		Title:       title,
		Url:         url,
		SubmittedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(submit); err != nil {
		s.Info("failed to submit content: %s\n", err)
	}
}

func (s *Shell) Comment(params []string) {
	if len(params) < 3 {
		s.Info("usage: comment <session-id> <item-id> <text>\n")
		return
	}
	sessionID, itemID, text := params[0], params[1], params[2]
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		s.Info("failed to find session %q: %s\n", sessionID, err)
		return
	}
	if q.Session == nil {
		s.Info("You are not logged in\n")
		return
	}
	commentOn := &PostComment{
		ParentID: NewTreeID(itemID),
		Author:   q.Session.Username,
		Content:  text,
		PostedAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(commentOn); err != nil {
		s.Info("failed to post comment: %s\n", err)
	}
}
