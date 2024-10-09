package main

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

type SubscriptionScope = string

var ErrInvalidSubscriptionScope = errors.New("invalid subscription scope")
var ErrSubscriptionSettingsNotFound = errors.New("subscription setting not found")

const (
	// When enabled, notifies the user about replies to their comments or submissions
	SUBSCRIPTION_SCOPE_REPLIES = "replies"
	// When enabled, notifies the user about new submissions
	SUBSCRIPTION_SCOPE_SUBMISSIONS = "submissions"
)

var AllowedSubscriptionScopes = []SubscriptionScope{
	SUBSCRIPTION_SCOPE_REPLIES,
	SUBSCRIPTION_SCOPE_SUBMISSIONS,
}

func ToSubscriptionScope(s string) (SubscriptionScope, error) {
	scopeIndex := slices.Index(AllowedSubscriptionScopes, s)

	if scopeIndex == -1 {
		return "", fmt.Errorf("%q not in %v: %w", s, AllowedSubscriptionScopes, ErrInvalidSubscriptionScope)
	}

	return AllowedSubscriptionScopes[scopeIndex], nil
}

func ScopeIs(s SubscriptionScope) func(s SubscriptionScope) bool {
	return func(other SubscriptionScope) bool {
		return s == other
	}
}

type SubscriptionSettings struct {
	LastChangeAt time.Time
	EnabledFor   []SubscriptionScope
	DisabledFor  []SubscriptionScope
	Subscriber   string
}

func NewDefaultSubscriptionSettings(username string, now time.Time) *SubscriptionSettings {
	return &SubscriptionSettings{
		LastChangeAt: now,
		EnabledFor:   []string{},
		DisabledFor:  slices.Clone(AllowedSubscriptionScopes),
		Subscriber:   username,
	}
}

func (s *SubscriptionSettings) Enable(cmd *EnableSubscriptions) {
	s.Subscriber = cmd.Username
	s.LastChangeAt = cmd.EnabledAt
	for _, scope := range cmd.Scopes {
		s.EnableScope(scope)
	}
}

func (s *SubscriptionSettings) Disable(cmd *DisableSubscriptions) {
	s.Subscriber = cmd.Username
	s.LastChangeAt = cmd.DisabledAt
	for _, scope := range cmd.Scopes {
		s.DisableScope(scope)
	}
}

func (s *SubscriptionSettings) EnableScope(scope SubscriptionScope) {
	if slices.Index(s.EnabledFor, scope) == -1 {
		s.EnabledFor = append(s.EnabledFor, scope)
	}
	s.DisabledFor = slices.DeleteFunc(s.DisabledFor, ScopeIs(scope))
}

func (s *SubscriptionSettings) DisableScope(scope SubscriptionScope) {
	if slices.Index(s.DisabledFor, scope) == -1 {
		s.DisabledFor = append(s.DisabledFor, scope)
	}
	s.EnabledFor = slices.DeleteFunc(s.EnabledFor, ScopeIs(scope))
}
