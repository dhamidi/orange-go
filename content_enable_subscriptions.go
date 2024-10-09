package main

import (
	"errors"
	"slices"
	"time"
)

type EnableSubscriptions struct {
	Username  string
	EnabledAt time.Time
	Scopes    []SubscriptionScope
}

func (cmd *EnableSubscriptions) CommandName() string { return "EnableSubscriptions" }

func init() {
	DefaultCommandRegistry.Register("EnableSubscriptions", func() Command { return new(EnableSubscriptions) })
}

func (self *Content) handleEnableSubscriptions(cmd *EnableSubscriptions) error {
	scopes := []SubscriptionScope{}
	for _, s := range cmd.Scopes {
		if scope, err := ToSubscriptionScope(s); err != nil {
			return err
		} else {
			scopes = append(scopes, scope)
		}
	}
	slices.Sort(scopes)
	record := &EnableSubscriptions{
		Username:  cmd.Username,
		EnabledAt: cmd.EnabledAt,
		Scopes:    slices.Compact(scopes),
	}

	// Persist the deduplicated and validated list of scopes only.
	*cmd = *record

	settings, err := self.state.GetSubscriptionSettings(cmd.Username)
	if errors.Is(err, ErrSubscriptionSettingsNotFound) {
		settings = NewDefaultSubscriptionSettings(cmd.Username, cmd.EnabledAt)
	} else if err != nil {
		return err
	}
	settings.Enable(record)
	return self.state.PutSubscriptionSettings(settings)
}
