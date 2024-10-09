package main

import (
	"errors"
	"slices"
	"time"
)

type DisableSubscriptions struct {
	Username   string
	DisabledAt time.Time
	Scopes     []SubscriptionScope
}

func (cmd *DisableSubscriptions) CommandName() string { return "DisableSubscriptions" }

func init() {
	DefaultCommandRegistry.Register("DisableSubscriptions", func() Command { return new(DisableSubscriptions) })
}

func (self *Content) handleDisableSubscriptions(cmd *DisableSubscriptions) error {
	scopes := []SubscriptionScope{}
	for _, s := range cmd.Scopes {
		if scope, err := ToSubscriptionScope(s); err != nil {
			return err
		} else {
			scopes = append(scopes, scope)
		}
	}
	slices.Sort(scopes)
	record := &DisableSubscriptions{
		Username:   cmd.Username,
		DisabledAt: cmd.DisabledAt,
		Scopes:     slices.Compact(scopes),
	}

	// Persist the deduplicated and validated list of scopes only.
	*cmd = *record

	settings, err := self.state.GetSubscriptionSettings(cmd.Username)
	if errors.Is(err, ErrSubscriptionSettingsNotFound) {
		settings = NewDefaultSubscriptionSettings(cmd.Username, cmd.DisabledAt)
	} else if err != nil {
		return err
	}
	settings.Disable(record)
	return self.state.PutSubscriptionSettings(settings)
}
