package main

import (
	"encoding/json"
	"fmt"
	"iter"
)

type Starter interface {
	Start() (stop func())
}

type Setupper interface {
	Setup() error
}

func MustSetup(s interface{}) {
	if setupper, ok := s.(Setupper); ok {
		if err := setupper.Setup(); err != nil {
			panic(fmt.Errorf("Failed to set up %v: %s", s, err))
		}
	}
}

type PersistedCommand struct {
	ID      int
	Message Command
}

type Command interface {
	CommandName() string
}

type Query interface {
	QueryName() string
}

type Serializer interface {
	Encode(message Command) ([]byte, error)
	Decode(data []byte, message *Command) error
}

type CommandHandler interface {
	HandleCommand(command Command) error
}

type QueryHandler interface {
	HandleQuery(query Query) error
}

var ErrCommandNotAccepted = fmt.Errorf("command not accepted")
var ErrQueryNotAccepted = fmt.Errorf("query not accepted")

type NewCommand func() Command
type JSONSerializer struct {
	messages map[string]NewCommand
}

type RawJSONMessage struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message"`
}

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{
		messages: make(map[string]NewCommand),
	}
}

func (s *JSONSerializer) Register(name string, newMessage NewCommand) {
	s.messages[name] = newMessage
}

func (s *JSONSerializer) Encode(message Command) ([]byte, error) {
	bytes, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	withType := RawJSONMessage{
		Type:    message.CommandName(),
		Message: bytes,
	}

	return json.Marshal(withType)
}

func (s *JSONSerializer) Decode(data []byte, message *Command) error {
	m := &RawJSONMessage{}
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	if newMessage, ok := s.messages[m.Type]; ok {
		*message = newMessage()
	}
	return json.Unmarshal(m.Message, message)
}

var DefaultSerializer = NewJSONSerializer()

type CommandLog interface {
	Append(command Command) error
	After(id int) (iter.Seq[*PersistedCommand], error)
}

// NullCommand log implements an empty log that does not store any messages nor return any.
type NullCommandLog struct{}

func (l *NullCommandLog) Append(command Command) error { return nil }
func (l *NullCommandLog) After(id int) (iter.Seq[*PersistedCommand], error) {
	return func(yield func(c *PersistedCommand) bool) {}, nil
}
