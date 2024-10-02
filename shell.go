package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type QueryBuilder = func(shell *Shell, req *Request, ctx context.Context) (Query, error)
type CommandBuilder = func(shell *Shell, req *Request, ctx context.Context) (Command, error)
type ContextBuilder = func(shell *Shell, req *Request, ctx context.Context) (context.Context, error)

type Shell struct {
	CurrentTime     func() time.Time
	NewID           func() string
	App             *App
	QueryBuilders   map[string]QueryBuilder
	CommandBuilders map[string]CommandBuilder
	ContextBuilders []ContextBuilder
}

var DefaultShellCommands = map[string]CommandBuilder{}
var DefaultShellQueries = map[string]QueryBuilder{}

func NewDefaultShell(app *App) *Shell {
	s := &Shell{
		CurrentTime:     time.Now,
		NewID:           uuid.NewString,
		App:             app,
		QueryBuilders:   map[string]QueryBuilder{},
		CommandBuilders: map[string]CommandBuilder{},
		ContextBuilders: []ContextBuilder{},
	}
	s.Use(CurrentTime).Use(CurrentSession)
	for name, builder := range DefaultShellCommands {
		s.RegisterCommand(name, builder)
	}
	for name, builder := range DefaultShellQueries {
		s.RegisterQuery(name, builder)
	}

	return s
}

func init() {
	DefaultShellCommands["PostLink"] = BuildPostLinkCommand
	DefaultShellCommands["Signup"] = BuildSignupCommand
	DefaultShellCommands["LogIn"] = BuildLoginCommand
	DefaultShellCommands["RequestMagicLinkLogin"] = BuildRequestMagicLinkLoginCommand
	DefaultShellCommands["LogInWithMagic"] = BuildLoginWithMagicCommand
	DefaultShellCommands["LinkVerifiedEmailToUser"] = BuildLinkVerifiedEmailToUserCommand
	DefaultShellCommands["SetAdminUsers"] = BuildSetAdminUsersCommand
	DefaultShellCommands["SetMagicDomains"] = BuildSetMagicDomainsCommand
	DefaultShellCommands["Upvote"] = BuildUpvoteCommand
	DefaultShellCommands["HideSubmission"] = BuildHideSubmissionCommand
	DefaultShellCommands["UnhideSubmission"] = BuildUnhideSubmissionCommand
	DefaultShellCommands["HideComment"] = BuildHideCommentCommand
	DefaultShellCommands["UnhideComment"] = BuildUnhideCommentCommand
	DefaultShellCommands["Hide"] = BuildHideCommentCommand
	DefaultShellCommands["Unhide"] = BuildUnhideCommentCommand
	DefaultShellCommands["PostLink"] = BuildPostLinkCommand
	DefaultShellCommands["RequestPasswordReset"] = BuildRequestPasswordResetCommand
	DefaultShellCommands["ResetPassword"] = BuildResetPasswordCommand
	DefaultShellCommands["Comment"] = BuildCommentCommand
	DefaultShellCommands["SetDefaultUsernamePolicy"] = BuildSetDefaultUsernamePolicyCommand
	DefaultShellCommands["SetChangeUsernamePolicy"] = BuildChangeUsernamePolicyCommand

	DefaultShellCommands["QueueEmail"] = BuildQueueEmailCommand
	DefaultShellCommands["SendWelcomeEmail"] = BuildSendWelcomeEmailCommand

	DefaultShellQueries["GetUserRoles"] = BuildGetUserRolesQuery
	DefaultShellQueries["FindSession"] = BuildFindSessionQuery
	DefaultShellQueries["GetFrontpage"] = BuildGetFrontpageQuery
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

func BuildSignupCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	createdAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("signup: %w", err)
	}
	username := req.Parameters.Get("username")
	password := req.Parameters.Get("password")
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("signup: failed to  hash password: %w", err)
	}
	return &SignUpUser{
		Username:     username,
		PasswordHash: *passwordHash,
		CreatedAt:    createdAt,
	}, nil
}

func BuildLoginCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	createdAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	username := req.Parameters.Get("username")
	password := req.Parameters.Get("password")
	sessionID := req.Parameters.Get("sessionID")
	if sessionID == "" {
		sessionID = shell.NewID()
	}
	q := NewFindUserPasswordHash(username, password)
	if err := shell.App.HandleQuery(q); err != nil {
		return nil, fmt.Errorf("login: failed to hash password: %w", err)
	}
	if q.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}
	return &LogInUser{
		Username:     username,
		PasswordHash: *q.PasswordHash,
		AttemptedAt:  createdAt,
		SessionID:    sessionID,
	}, nil
}

func BuildRequestMagicLinkLoginCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	createdAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("request-magic-link-login: %w", err)
	}
	email := req.Parameters.Get("email")
	magic := shell.NewID()
	return &RequestMagicLinkLogin{
		Email:       email,
		Magic:       magic,
		RequestedAt: createdAt,
	}, nil
}

func BuildLoginWithMagicCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	createdAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("login-with-magic: %w", err)
	}
	magic := req.Parameters.Get("magic")
	return &LogInWithMagic{
		SessionID:   req.Parameters.Get("sessionID"),
		Magic:       magic,
		AttemptedAt: createdAt,
	}, nil
}

func BuildLinkVerifiedEmailToUserCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	createdAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("link-verified-email-to-user: %w", err)
	}
	username := req.Parameters.Get("username")
	email := req.Parameters.Get("email")
	return &LinkVerifiedEmailToUser{
		Username: username,
		Email:    email,
		LinkedAt: createdAt,
	}, nil
}

func BuildSetAdminUsersCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	usernames := GetAllValues(req.Parameters, "username")
	return &SetAdminUsers{Users: usernames}, nil
}

func BuildSetMagicDomainsCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	domains := GetAllValues(req.Parameters, "domain")
	return &SetMagicDomains{Domains: domains}, nil
}

func BuildGetUserRolesQuery(shell *Shell, req *Request, ctx context.Context) (Query, error) {
	username := req.Parameters.Get("username")
	return NewGetUserRolesQuery(username), nil
}

func BuildFindSessionQuery(shell *Shell, req *Request, ctx context.Context) (Query, error) {
	sessionID := req.Parameters.Get("sessionID")
	return NewFindSessionQuery(sessionID), nil
}

func BuildGetFrontpageQuery(shell *Shell, req *Request, ctx context.Context) (Query, error) {
	env := NewRequestEnv(ctx)
	var viewer string
	if session := env.CurrentSession(); session != nil {
		viewer = session.Username
	}
	return NewFrontpageQuery(&viewer), nil
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

func BuildUpvoteCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	votedAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("upvote: %w", err)
	}
	session := env.CurrentSession()
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return &UpvoteSubmission{
		ItemID:  req.Parameters.Get("itemID"),
		Voter:   session.Username,
		VotedAt: votedAt,
	}, nil
}

func BuildHideSubmissionCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	hiddenAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("hide-submission: %w", err)
	}
	session := env.CurrentSession()
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return &HideSubmission{
		ItemID:   req.Parameters.Get("itemID"),
		HiddenBy: session.Username,
		HiddenAt: hiddenAt,
	}, nil
}

func BuildUnhideSubmissionCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	unhiddenAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("unhide-submission: %w", err)
	}
	session := env.CurrentSession()
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return &UnhideSubmission{
		ItemID:     req.Parameters.Get("itemID"),
		UnhiddenBy: session.Username,
		UnhiddenAt: unhiddenAt,
	}, nil
}

func BuildHideCommentCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	hiddenAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("hide-comment: %w", err)
	}
	session := env.CurrentSession()
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return &HideComment{
		CommentID: NewTreeID(req.Parameters.Get("itemID")),
		HiddenBy:  session.Username,
		HiddenAt:  hiddenAt,
	}, nil
}

func BuildUnhideCommentCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	unhiddenAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("unhide-comment: %w", err)
	}
	session := env.CurrentSession()
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return &UnhideComment{
		CommentID:  NewTreeID(req.Parameters.Get("itemID")),
		UnhiddenBy: session.Username,
		UnhiddenAt: unhiddenAt,
	}, nil
}

func BuildPostLinkCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	submittedAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("post-link: %w", err)
	}
	session := env.CurrentSession()
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return &PostLink{
		ItemID:      req.Parameters.Get("itemID"),
		Submitter:   session.Username,
		Title:       req.Parameters.Get("title"),
		Url:         req.Parameters.Get("url"),
		SubmittedAt: submittedAt,
	}, nil
}

func BuildRequestPasswordResetCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	requestedAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("request-password-reset: %w", err)
	}
	email := req.Parameters.Get("email")
	q := NewFindUserByEmail(email)
	if err := shell.App.HandleQuery(q); err != nil {
		return nil, fmt.Errorf("request-password-reset: %w", err)
	}
	if q.User == nil {
		return nil, ErrUserNotFound
	}
	return &RequestPasswordReset{
		Username:    q.User.Username,
		Email:       email,
		Token:       shell.NewID(),
		RequestedAt: requestedAt,
	}, nil
}

func BuildResetPasswordCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	requestedAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("reset-password: %w", err)
	}
	password, err := HashPassword(req.Parameters.Get("password"))
	if err != nil {
		return nil, fmt.Errorf("reset-password: failed to hash password: %w", err)
	}
	return &ResetPassword{
		Token:       req.Parameters.Get("token"),
		NewPassword: *password,
		AttemptedAt: requestedAt,
	}, nil
}

func BuildCommentCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	postedAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("comment: %w", err)
	}
	session := env.CurrentSession()
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return &PostComment{
		ParentID: NewTreeID(req.Parameters.Get("itemID")),
		Author:   session.Username,
		Content:  req.Parameters.Get("text"),
		PostedAt: postedAt,
	}, nil
}

func BuildSetDefaultUsernamePolicyCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	return &ChangeUsernamePolicy{
		MinLength: 5,
		MaxLength: 100,
		Blacklist: []string{"guest"},
	}, nil
}

func BuildChangeUsernamePolicyCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	minLength, err := strconv.Atoi(req.Parameters.Get("minLength"))
	if err != nil {
		return nil, fmt.Errorf("change-username-policy: failed to parse minLength: %w", err)
	}
	maxLength, err := strconv.Atoi(req.Parameters.Get("maxLength"))
	if err != nil {
		return nil, fmt.Errorf("change-username-policy: failed to parse maxLength: %w", err)
	}
	blacklist := GetAllValues(req.Parameters, "blacklist")

	return &ChangeUsernamePolicy{
		MinLength: minLength,
		MaxLength: maxLength,
		Blacklist: blacklist,
	}, nil
}

func BuildQueueEmailCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	templateData := map[string]any{}
	if err := json.Unmarshal([]byte(req.Parameters.Get("templateData")), &templateData); err != nil {
		return nil, fmt.Errorf("queue-email: failed to decode templateData: %w", err)
	}
	internalID := req.Parameters.Get("internalID")
	if internalID == "" {
		internalID = shell.NewID()
	}
	return &QueueEmail{
		InternalID:   internalID,
		Recipients:   req.Parameters.Get("email"),
		Subject:      req.Parameters.Get("subject"),
		TemplateName: req.Parameters.Get("template"),
		TemplateData: templateData,
	}, nil
}

func BuildSendWelcomeEmailCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	p := req.Parameters
	internalID := req.Parameters.Get("internalID")
	if internalID == "" {
		internalID = shell.NewID()
	}
	email := strings.TrimSpace(p.Get("email"))
	if email == "" {
		return nil, fmt.Errorf("send-welcome-email: no email provided")
	}

	name := strings.TrimSpace(p.Get("username"))
	if name == "" {
		return nil, fmt.Errorf("send-welcome-email: no username provided")
	}

	cmd := &QueueEmail{
		InternalID:   internalID,
		Recipients:   email,
		Subject:      "",
		TemplateName: "welcome",
		TemplateData: map[string]any{
			"name":      name,
			"username":  name,
			"login_url": "https://orange.decode.ee/login",
		},
	}
	return cmd, nil
}

var ErrUnkownRequestKind = errors.New("unknown request kind")

func (s *Shell) Do(ctx context.Context, req *Request) (any, error) {
	kind := req.Kind()
	if kind == UnknownRequest {
		return nil, ErrUnkownRequestKind
	}

	if kind == CommandRequest {
		command, err := s.buildCommand(req, ctx)
		if err != nil {
			return nil, err
		}
		return nil, s.App.HandleCommand(command)
	}

	if kind == QueryRequest {
		query, err := s.buildQuery(req, ctx)
		if err != nil {
			return nil, err
		}
		if err := s.App.HandleQuery(query); err != nil {
			return nil, err
		}
		return query.Result(), nil
	}
	panic("unreachable")
}

func (s *Shell) buildCommand(req *Request, ctx context.Context) (Command, error) {
	name := req.Name()
	commandBuilder, ok := s.CommandBuilders[name]
	if !ok {
		return nil, fmt.Errorf("BuildCommand(%q): %w", name, ErrCommandNotAccepted)
	}
	enhancedCtx, err := s.buildContext(req, ctx)
	if err != nil {
		return nil, fmt.Errorf("BuildCommand(%q): %w", name, err)
	}
	return commandBuilder(s, req, enhancedCtx)
}

func (s *Shell) buildQuery(req *Request, ctx context.Context) (Query, error) {
	name := req.Name()
	queryBuilder, ok := s.QueryBuilders[name]
	if !ok {
		return nil, fmt.Errorf("BuildQuery(%q): %w", name, ErrQueryNotAccepted)
	}
	enhancedCtx, err := s.buildContext(req, ctx)
	if err != nil {
		return nil, fmt.Errorf("BuildQuery(%q): %w", name, err)
	}
	return queryBuilder(s, req, enhancedCtx)
}

func (s *Shell) buildContext(req *Request, ctx context.Context) (context.Context, error) {
	enhancedCtx := ctx
	for _, builder := range s.ContextBuilders {
		var err error
		enhancedCtx, err = builder(s, req, enhancedCtx)
		if err != nil {
			return nil, err
		}
	}
	return enhancedCtx, nil
}

func (s *Shell) RegisterCommand(name string, builder CommandBuilder) *Shell {
	s.CommandBuilders[name] = builder
	return s
}
func (s *Shell) RegisterQuery(name string, builder QueryBuilder) *Shell {
	s.QueryBuilders[name] = builder
	return s
}

func (s *Shell) Use(ctxb ContextBuilder) *Shell {
	s.ContextBuilders = append(s.ContextBuilders, ctxb)
	return s
}

type requestEnv = int

var ErrNotInEnv = errors.New("not in environment")

const (
	EnvCurrentTime requestEnv = iota
	EnvCurrentSession
)

type RequestEnv struct{ context.Context }

func NewRequestEnv(ctx context.Context) RequestEnv {
	return RequestEnv{ctx}
}

func (e *RequestEnv) CurrentTime() (time.Time, error) {
	return CurrentTimeFromEnv(e.Context)
}
func (e *RequestEnv) CurrentSession() *Session {
	return CurrentSessionFromEnv(e.Context)
}

// CurrentTime adds `EnvCurrentTime` to the context with the current time.
func CurrentTime(shell *Shell, req *Request, ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, EnvCurrentTime, shell.CurrentTime()), nil
}

// CurrentTimeFromEnv returns the current time from the context.
func CurrentTimeFromEnv(ctx context.Context) (time.Time, error) {
	v := ctx.Value(EnvCurrentTime)
	if asTime, ok := v.(time.Time); ok {
		return asTime, nil
	}
	return time.Time{}, ErrNotInEnv
}

// CurrentSession adds `EnvCurrentSession` to the context with the current session.
//
// The session is determined by a header called `sessionID`.
func CurrentSession(shell *Shell, req *Request, ctx context.Context) (context.Context, error) {
	sessionID := req.Headers.Get("sessionID")
	if sessionID == "" {
		sessionID = req.Parameters.Get("sessionID")
	}
	q := NewFindSessionQuery(sessionID)
	if err := shell.App.HandleQuery(q); err != nil {
		return nil, fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	session := q.Session
	if session == nil {
		return ctx, nil
	}

	return context.WithValue(ctx, EnvCurrentSession, session), nil
}

// CurrentSessionFromEnv returns the current session from the context.
// If no session is found, it returns nil.
func CurrentSessionFromEnv(ctx context.Context) *Session {
	v := ctx.Value(EnvCurrentSession)
	if asSession, ok := v.(*Session); ok {
		return asSession
	}
	return nil
}
