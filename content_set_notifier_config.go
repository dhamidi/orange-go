package main

import (
	"time"
)

type SetNotifierConfig struct {
	Enabled   bool
	ChangedAt time.Time
}

func (cmd *SetNotifierConfig) CommandName() string { return "SetNotifierConfig" }

func init() {
	DefaultCommandRegistry.Register("SetNotifierConfig", func() Command { return new(SetNotifierConfig) })
}

func (self *Content) handleSetNotifierConfig(cmd *SetNotifierConfig) error {
	return nil
}
