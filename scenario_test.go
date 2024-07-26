package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type TestContext struct {
	t         *testing.T
	App       *App
	Starters  []Starter
	PostIDs   []string
	Submitter string
	Viewer    string
}

func (t *TestContext) upvote(itemID, as string) Command {
	return &UpvoteSubmission{
		ItemID:  itemID,
		Voter:   as,
		VotedAt: time.Now(),
	}
}

func (t *TestContext) postLink(url, title string) Command {
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

func (t *TestContext) frontpage() []*Submission {
	t.t.Helper()
	top := NewFrontpageQuery(&t.Viewer)
	if err := t.App.HandleQuery(top); err != nil {
		t.t.Fatalf("%s", err)
	}
	return top.Submissions
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
