package main

import "iter"

type InMemoryCommandLog struct {
	messages []*PersistedCommand
}

func (self *InMemoryCommandLog) After(id int) (iter.Seq[*PersistedCommand], error) {
	if len(self.messages) < id {
		id = len(self.messages) - 1
	}
	messages := self.messages[id:]
	after := func(yield func(*PersistedCommand) bool) {
		for _, c := range messages {
			if yield(c) == false {
				return
			}
		}
	}
	return after, nil
}

// Append implements CommandLog.
func (self *InMemoryCommandLog) Append(command Command) error {
	id := len(self.messages)
	id += 1
	newEntry := &PersistedCommand{ID: id, Message: command}
	self.messages = append(self.messages, newEntry)
	return nil
}

func NewInMemoryCommandLog() *InMemoryCommandLog {
	return &InMemoryCommandLog{
		messages: []*PersistedCommand{},
	}
}
