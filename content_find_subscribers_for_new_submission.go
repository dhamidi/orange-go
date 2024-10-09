package main

import "fmt"

type FindSubscribersForNewSubmission struct {
	Subscribers []string
}

func (q *FindSubscribersForNewSubmission) QueryName() string {
	return "FindSubscribersForNewSubmission"
}

func (q *FindSubscribersForNewSubmission) Result() any { return q.Subscribers }

func NewFindSubscribersForNewSubmission() *FindSubscribersForNewSubmission {
	return &FindSubscribersForNewSubmission{
		Subscribers: []string{},
	}
}

func (self *Content) findSubscribersForNewSubmission(q *FindSubscribersForNewSubmission) error {
	subscribers, err := self.state.GetActiveSubscribers()
	if err != nil {
		return err
	}

	for _, username := range subscribers {
		settings, err := self.state.GetSubscriptionSettings(username)
		if err != nil {
			return fmt.Errorf("failed to fetch subscription settings for %q: %w", username, err)
		}
		if settings.HasScope(SUBSCRIPTION_SCOPE_SUBMISSIONS) {
			q.Subscribers = append(q.Subscribers, username)
		}
	}
	return nil
}
