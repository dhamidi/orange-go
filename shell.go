package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
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
	DefaultShellCommands["Login"] = BuildLoginCommand
	DefaultShellCommands["RequestMagicLinkLogin"] = BuildRequestMagicLinkLoginCommand
	DefaultShellCommands["LoginWithMagicLink"] = BuildLoginWithMagicLinkCommand
	DefaultShellCommands["LinkVerifiedEmailToUser"] = BuildLinkVerifiedEmailToUserCommand
	DefaultShellCommands["SetAdminUsers"] = BuildSetAdminUsersCommand
	DefaultShellCommands["SetMagicDomains"] = BuildSetMagicDomainsCommand
	DefaultShellCommands["Upvote"] = BuildUpvoteCommand
	DefaultShellCommands["HideSubmission"] = BuildHideSubmissionCommand
	DefaultShellCommands["UnhideSubmission"] = BuildUnhideSubmissionCommand
	DefaultShellCommands["Hide"] = BuildHideCommentCommand
	DefaultShellCommands["Unhide"] = BuildUnhideCommentCommand
	DefaultShellCommands["PostLink"] = BuildPostLinkCommand
	DefaultShellCommands["RequestPasswordReset"] = BuildRequestPasswordResetCommand
	DefaultShellCommands["ResetPassword"] = BuildResetPasswordCommand
	DefaultShellCommands["Comment"] = BuildCommentCommand

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

func BuildLoginCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	createdAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	username := req.Parameters.Get("username")
	password := req.Parameters.Get("password")
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
		SessionID:    shell.NewID(),
	}, nil
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

func BuildLoginWithMagicLinkCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	env := NewRequestEnv(ctx)
	createdAt, err := env.CurrentTime()
	if err != nil {
		return nil, fmt.Errorf("login-with-magic-link: %w", err)
	}
	magic := req.Parameters.Get("magic")
	sessionID := shell.NewID()
	return &LogInWithMagic{
		SessionID:   sessionID,
		Magic:       magic,
		AttemptedAt: createdAt,
	}, nil
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

func (s *Shell) SetAdminUsers(params Parameters) error {
	usernames := GetAllValues(params, "username")
	setAdmin := &SetAdminUsers{Users: usernames}
	if err := s.App.HandleCommand(setAdmin); err != nil {
		return fmt.Errorf("set-admin-users: %w\n", err)
	}
	return nil
}

func BuildSetAdminUsersCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	usernames := GetAllValues(req.Parameters, "username")
	return &SetAdminUsers{Users: usernames}, nil
}

func (s *Shell) SetMagicDomains(params Parameters) error {
	domains := GetAllValues(params, "domain")
	setMagicDomains := &SetMagicDomains{Domains: domains}
	if err := s.App.HandleCommand(setMagicDomains); err != nil {
		return fmt.Errorf("set-magic-domains: %w\n", err)
	}
	return nil
}

func BuildSetMagicDomainsCommand(shell *Shell, req *Request, ctx context.Context) (Command, error) {
	domains := GetAllValues(req.Parameters, "domain")
	return &SetMagicDomains{Domains: domains}, nil
}

func (s *Shell) GetUserRoles(params Parameters) ([]UserRole, error) {
	username := params.Get("username")
	q := NewGetUserRolesQuery(username)
	if err := s.App.HandleQuery(q); err != nil {
		return nil, fmt.Errorf("get-user-roles: failed to get user roles: %w\n", err)
	}
	return q.Roles, nil
}

func BuildGetUserRolesQuery(shell *Shell, req *Request, ctx context.Context) (Query, error) {
	username := req.Parameters.Get("username")
	return NewGetUserRolesQuery(username), nil
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

func BuildFindSessionQuery(shell *Shell, req *Request, ctx context.Context) (Query, error) {
	sessionID := req.Parameters.Get("sessionID")
	return NewFindSessionQuery(sessionID), nil
}

func (s *Shell) GetFrontpage(params Parameters) ([]*Submission, error) {
	sessionID := params.Get("sessionID")
	var viewer string
	if sessionID != "" {
		q := NewFindSessionQuery(sessionID)
		if err := s.App.HandleQuery(q); err != nil {
			return nil, fmt.Errorf("get-frontpage: failed to find session %q: %w\n", sessionID, err)
		}
		if q.Session == nil {
			return nil, ErrSessionNotFound
		}
		viewer = q.Session.Username
	}
	frontpage := NewFrontpageQuery(&viewer)
	if err := s.App.HandleQuery(frontpage); err != nil {
		return nil, fmt.Errorf("get-frontpage: %w\n", err)
	}
	return frontpage.Submissions, nil
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

func (s *Shell) HideSubmission(params Parameters) error {
	sessionID := params.Get("sessionID")
	itemID := params.Get("itemID")
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return ErrSessionNotFound
	}

	submit := &HideSubmission{
		ItemID:   itemID,
		HiddenBy: q.Session.Username,
		HiddenAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(submit); err != nil {
		return fmt.Errorf("hide-submission: %w", err)
	}
	return nil
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

func (s *Shell) UnhideSubmission(params Parameters) error {
	sessionID := params.Get("sessionID")
	itemID := params.Get("itemID")
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return ErrSessionNotFound
	}
	submit := &UnhideSubmission{
		ItemID:     itemID,
		UnhiddenBy: q.Session.Username,
		UnhiddenAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(submit); err != nil {
		return fmt.Errorf("hide-submission: %w", err)
	}
	return nil
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

func (s *Shell) HideComment(params Parameters) error {
	sessionID := params.Get("sessionID")
	itemID := params.Get("itemID")
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return ErrSessionNotFound
	}

	submit := &HideComment{
		CommentID: NewTreeID(itemID),
		HiddenBy:  q.Session.Username,
		HiddenAt:  s.CurrentTime(),
	}
	if err := s.App.HandleCommand(submit); err != nil {
		return fmt.Errorf("hide-comment: %w", err)
	}
	return nil
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

func (s *Shell) UnhideComment(params Parameters) error {
	sessionID := params.Get("sessionID")
	itemID := params.Get("itemID")
	q := NewFindSessionQuery(sessionID)
	if err := s.App.HandleQuery(q); err != nil {
		return fmt.Errorf("failed to find session %q: %w\n", sessionID, err)
	}
	if q.Session == nil {
		return ErrSessionNotFound
	}
	unhide := &UnhideComment{
		CommentID:  NewTreeID(itemID),
		UnhiddenBy: q.Session.Username,
		UnhiddenAt: s.CurrentTime(),
	}
	if err := s.App.HandleCommand(unhide); err != nil {
		return fmt.Errorf("unhide-comment: %w", err)
	}
	return nil
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
