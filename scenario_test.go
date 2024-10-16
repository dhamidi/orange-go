package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

type TestContext struct {
	t         *testing.T
	App       *App
	Starters  []Starter
	PostIDs   []string
	Submitter string
	Viewer    string
}

func (t *TestContext) Notifier() *Notifier {
	for _, s := range t.Starters {
		notifier, ok := s.(*Notifier)
		if ok {
			return notifier
		}
	}
	return nil
}

func (t *TestContext) upvote(itemID, as string) Command {
	return &UpvoteSubmission{
		ItemID:  itemID,
		Voter:   as,
		VotedAt: time.Now(),
	}
}
func (t *TestContext) upvoteN(itemID string, n int) {
	for i := 0; i < n; i++ {
		t.must(t.upvote(itemID, fmt.Sprintf("viewer-%d", i)))
	}
}

func (t *TestContext) postLink(url, title string) *PostLink {
	itemID := fmt.Sprintf("post-%d", len(t.PostIDs)+1)
	t.PostIDs = append(t.PostIDs, itemID)
	return &PostLink{
		ItemID:      itemID,
		Submitter:   t.Submitter,
		Url:         url,
		Title:       title,
		SubmittedAt: time.Now(),
	}
}

func (t *TestContext) commentOn(itemID string, content string) Command {
	return &PostComment{
		ParentID: NewTreeID(itemID),
		Content:  content,
		Author:   t.Viewer,
		PostedAt: time.Now(),
	}
}

func (t *TestContext) signup(username, password string) Command {
	t.t.Helper()
	passwordHash, err := HashPassword(password)
	if err != nil {
		t.t.Fatalf("failed to hash password: %s", err)
	}
	return &SignUpUser{
		Username:     username,
		PasswordHash: *passwordHash,
		CreatedAt:    time.Now(),
	}
}

func (t *TestContext) login(username string, password string) Command {
	t.t.Helper()
	sessionID := uuid.NewString()
	q := NewFindUserPasswordHash(username, password)
	if err := t.App.HandleQuery(q); err != nil {
		t.t.Fatalf("failed to find password hash: %s", err)
	}
	return &LogInUser{
		SessionID:    sessionID,
		Username:     username,
		PasswordHash: *q.PasswordHash,
		AttemptedAt:  time.Now(),
	}
}

func (t *TestContext) frontpage() []*Submission {
	t.t.Helper()
	return t.frontpageAfter(0)
}

func (t *TestContext) frontpageAfter(after int) []*Submission {
	t.t.Helper()
	top := NewFrontpageQuery(&t.Viewer)
	top.After = after
	if err := t.App.HandleQuery(top); err != nil {
		t.t.Fatalf("%s", err)
	}
	return top.Submissions
}

func (t *TestContext) findPasswordHash(username, password string) (*PasswordHash, error) {
	t.t.Helper()
	q := NewFindUserPasswordHash(username, password)
	if err := t.App.HandleQuery(q); err != nil {
		return nil, err
	}
	return q.PasswordHash, nil
}

func (t *TestContext) findUserByEmail(email string) (*User, error) {
	t.t.Helper()
	q := NewFindUserByEmail(email)
	if err := t.App.HandleQuery(q); err != nil {
		return nil, err
	}
	return q.User, nil
}

func (t *TestContext) setUsernamePolicy(minLength, maxLength int) Command {
	return &ChangeUsernamePolicy{
		MinLength: minLength,
		MaxLength: maxLength,
	}
}

func (t *TestContext) forbidUsername(usernames ...string) Command {
	return &ChangeUsernamePolicy{
		MinLength: 0,
		MaxLength: 100,
		Blacklist: usernames,
	}
}

func (t *TestContext) linkVerifiedEmailToUser(username string, email string) Command {
	return &LinkVerifiedEmailToUser{
		Username: username,
		Email:    email,
		LinkedAt: time.Now(),
	}
}

func (t *TestContext) resetPassword(token, password string) Command {
	t.t.Helper()
	passwordHash, err := HashPassword(password)
	if err != nil {
		t.t.Fatalf("failed to hash password: %s", err)
	}
	return &ResetPassword{
		Token:       token,
		NewPassword: *passwordHash,
		AttemptedAt: time.Now(),
	}
}

func (t *TestContext) requestPasswordReset(username string, email string) Command {
	return &RequestPasswordReset{
		Username:    username,
		Email:       email,
		Token:       uuid.NewString(),
		RequestedAt: time.Now(),
	}
}

func (t *TestContext) requestMagicLinkLogin(email, magic string) Command {
	return &RequestMagicLinkLogin{
		Email:       email,
		Magic:       magic,
		RequestedAt: time.Now(),
	}
}

func (t *TestContext) loginWithMagic(magic string) Command {
	return &LogInWithMagic{
		SessionID:   uuid.NewString(),
		Magic:       magic,
		AttemptedAt: time.Now(),
	}
}

func (t *TestContext) setMagicDomains(domains ...string) Command {
	return &SetMagicDomains{
		Domains: domains,
	}
}

func (t *TestContext) hideSubmission(id string) Command {
	return &HideSubmission{
		ItemID:   id,
		HiddenBy: t.Viewer,
		HiddenAt: time.Now(),
	}
}

func (t *TestContext) unhideSubmission(id string) Command {
	return &UnhideSubmission{
		ItemID:     id,
		UnhiddenBy: t.Viewer,
		UnhiddenAt: time.Now(),
	}
}

func (t *TestContext) subscribeTo(username string, scope SubscriptionScope) Command {
	return &EnableSubscriptions{
		Username:  username,
		EnabledAt: time.Now(),
		Scopes:    []string{scope},
	}
}

func (t *TestContext) enableNotifier() Command {
	return &SetNotifierConfig{
		ChangedAt: time.Now(),
		Enabled:   true,
	}
}

func (t *TestContext) do(cmd Command) error {
	return t.App.HandleCommand(cmd)
}

func (t *TestContext) must(cmd Command) {
	t.t.Helper()
	if err := t.App.HandleCommand(cmd); err != nil {
		t.t.Fatalf("failed to execute %s: %s", cmd.CommandName(), err)
	}
}

func (t *TestContext) mustFailWith(cmd Command, expected error) {
	t.t.Helper()
	err := t.do(cmd)
	if err == nil {
		t.t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, expected) {
		t.t.Fatalf("expected %s, got %s", expected, err)
	}
}

func (t *TestContext) LogContains(f func(*PersistedCommand) bool) bool {
	commands, err := t.App.Commands.After(0)
	if err != nil {
		t.t.Fatalf("Failed to read log: %s", err)
	}
	for command := range commands {
		if f(command) {
			return true
		}
	}
	return false
}

func setup(t *testing.T) *TestContext {
	config := NewPlatformConfigForTest()
	app, starters := HackerNews(config)
	return &TestContext{
		t:         t,
		App:       app,
		Starters:  starters,
		PostIDs:   []string{},
		Submitter: "test-user",
		Viewer:    "viewer",
	}
}

func mustFind[E any](collection []E, field string, value interface{}) E {
	for _, elem := range collection {
		if reflect.ValueOf(elem).Elem().FieldByName(field).Interface() == value {
			return elem
		}
	}

	panic(fmt.Sprintf("failed to find %s with %s=%v", reflect.TypeOf(collection).Elem().String(), field, value))
}

func (t *TestContext) DumpLog() {
	commands, err := t.App.Commands.After(0)
	if err != nil {
		t.t.Fatalf("DumpLog: %s", err)
	}
	for command := range commands {
		t.t.Logf("% 3d %s %#v", command.ID, command.Message.CommandName(), command.Message)
	}
}
