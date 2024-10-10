package main

import (
	"fmt"
	"slices"
)

type FindSubscribersForNewComment struct {
	ParentID    TreeID
	Subscribers []string
}

func (q *FindSubscribersForNewComment) QueryName() string {
	return "FindSubscribersForNewComment"
}

func (q *FindSubscribersForNewComment) Result() any { return q.Subscribers }

func NewFindSubscribersForNewComment(parentID string) *FindSubscribersForNewComment {
	return &FindSubscribersForNewComment{
		ParentID:    NewTreeID(parentID),
		Subscribers: []string{},
	}
}

func (self *Content) findSubscribersForNewComment(q *FindSubscribersForNewComment) error {
	submission, err := self.state.GetSubmission(q.ParentID.Root())
	if err != nil {
		return fmt.Errorf("could not find root submission for %q: %w", q.ParentID, err)
	}
	result := map[string]struct{}{submission.Submitter: {}}
	var walk func(comment *Comment)
	walk = func(comment *Comment) {
		_, recorded := result[comment.Author]
		if !recorded {
			result[comment.Author] = struct{}{}
		}
		for _, c := range comment.Children {
			walk(c)
		}
	}
	for _, comment := range submission.Comments {
		walk(comment)
	}
	activeSubscribers, err := self.state.GetActiveSubscribers()
	if err != nil {
		return fmt.Errorf("FindSubscribersForNewComment: failed to get active subscribers: %w", err)
	}
	for username := range result {
		if slices.Contains(activeSubscribers, username) {
			q.Subscribers = append(q.Subscribers, username)
		}
	}
	return nil
}
