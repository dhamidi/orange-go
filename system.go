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

type SkipCommand struct{}

func (cmd *SkipCommand) CommandName() string { return "SkipCommand" }

type SkipHandler struct{}

func (self *SkipHandler) HandleCommand(cmd Command) error {
	switch cmd.(type) {
	case *SkipCommand:
		return nil
	default:
		return ErrCommandNotAccepted
	}
}

func init() {
	DefaultCommandRegistry.Register("SkipCommand", func() Command { return &SkipCommand{} })
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

type CommandRegistry map[string]NewCommand

func (s CommandRegistry) Register(name string, newMessage NewCommand) {
	s[name] = newMessage
}

func (s CommandRegistry) New(name string) (Command, bool) {
	newMessage, ok := s[name]
	if !ok {
		return nil, false
	}
	return newMessage(), true
}

var DefaultCommandRegistry = make(CommandRegistry)

type NewCommand func() Command
type JSONSerializer struct {
	commands CommandRegistry
}

type RawJSONMessage struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message"`
}

func NewJSONSerializer(registry CommandRegistry) *JSONSerializer {
	return &JSONSerializer{
		registry,
	}
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
	if newMessage, ok := s.commands.New(m.Type); ok {
		*message = newMessage
	}
	return json.Unmarshal(m.Message, message)
}

var DefaultSerializer = NewJSONSerializer(DefaultCommandRegistry)

type CommandLog interface {
	Append(command Command) error
	After(id int) (iter.Seq[*PersistedCommand], error)
}

type CommandReviser interface {
	ReviseCommands(id []int, as func(id int) Command) error
}

// NullCommand log implements an empty log that does not store any messages nor return any.
type NullCommandLog struct{}

func (l *NullCommandLog) Append(command Command) error { return nil }
func (l *NullCommandLog) After(id int) (iter.Seq[*PersistedCommand], error) {
	return func(yield func(c *PersistedCommand) bool) {}, nil
}
