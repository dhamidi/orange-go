package main

import (
	"bytes"
	"encoding/json"
)

type MailMessage struct {
	Recipient    string
	TemplateName string
	TemplateData json.RawMessage // ???
}

type MailAgent interface {
	SendEmail(message *MailMessage) error
}

type MailAgentStub struct {
	sink *bytes.Buffer
}

func NewMailAgentStub() *MailAgentStub {
	buf := new(bytes.Buffer)
	return &MailAgentStub{sink: buf}
}

func (m *MailAgentStub) SendEmail(message *MailMessage) error {
	panic("implement me")
}
