package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Email struct {
	InternalID   string
	Recipient    string
	Subject      string
	TemplateName string
	TemplateData map[string]any

	Retries int
	Status  string
	Message string
}

type EmailSender interface {
	SendEmail(email *Email) (EmailReceipt, error)
}

type EmailReceipt interface {
	ExternalMessageID() string
}

type EmailLogger struct {
	Logger *log.Logger
}

func (logger *EmailLogger) SendEmail(email *Email) (EmailReceipt, error) {
	label := email.Subject
	if email.Subject == "" {
		label = email.TemplateName
	}
	templateData, _ := json.Marshal(email.TemplateData)
	logger.Logger.Printf("sending email to %s: %s %s", email.Recipient, label, string(templateData))
	return nil, nil
}

type QueueEmail struct {
	InternalID   string
	Recipients   string
	Subject      string
	TemplateName string
	TemplateData map[string]any
}

func (cmd *QueueEmail) CommandName() string {
	return "QueueEmail"
}

func init() {
	DefaultCommandRegistry.Register("QueueEmail", func() Command { return &QueueEmail{} })
}

type SetEmailDeliveryStatus struct {
	InternalID string
	Status     string
	Message    string
}

func (cmd *SetEmailDeliveryStatus) CommandName() string {
	return "SetEmailDeliveryStatus"
}

func init() {
	DefaultCommandRegistry.Register("SetEmailDeliveryStatus", func() Command { return &SetEmailDeliveryStatus{} })
}

const StatusDelivered = "delivered"
const StatusFailed = "failed"
const StatusQueued = "queued"

type Mailer struct {
	Sender  EmailSender
	Outbox  map[string]map[string]*Email
	Logger  *log.Logger
	App     *App
	Version int
}

func NewMailer(sender EmailSender, logger *log.Logger, app *App) *Mailer {
	return &Mailer{
		Sender: sender,
		Outbox: map[string]map[string]*Email{
			StatusQueued:    {},
			StatusDelivered: {},
			StatusFailed:    {},
		},
		Logger: logger,
		App:    app,
	}
}

func (self *Mailer) HandleCommand(cmd Command) error {
	switch cmd := cmd.(type) {
	case *QueueEmail:
		return self.addToOutbox(cmd)
	case *SetEmailDeliveryStatus:
		return self.removeFromOutbox(cmd)
	default:
		return ErrCommandNotAccepted
	}
}

func (self *Mailer) addToOutbox(cmd *QueueEmail) error {
	email := &Email{
		InternalID:   cmd.InternalID,
		Recipient:    cmd.Recipients,
		Subject:      cmd.Subject,
		TemplateName: cmd.TemplateName,
		TemplateData: cmd.TemplateData,
	}
	byStatus := self.Outbox[StatusQueued]
	byStatus[email.InternalID] = email
	return nil
}
func (self *Mailer) findMessageByID(id string) *Email {
	for _, messages := range self.Outbox {
		if message, ok := messages[id]; ok {
			return message
		}
	}
	return nil
}

func (self *Mailer) removeFromOutbox(cmd *SetEmailDeliveryStatus) error {
	message := self.findMessageByID(cmd.InternalID)
	if message == nil {
		return nil
	}
	for _, messages := range self.Outbox {
		delete(messages, cmd.InternalID)
	}

	byStatus := self.Outbox[cmd.Status]
	byStatus[cmd.InternalID] = message
	message.Status = cmd.Status
	return nil
}
func (self *Mailer) Start() func() {
	stop := make(chan struct{})
	self.catchUp()
	self.logOutbox()
	go self.loop(stop)
	return func() { stop <- struct{}{} }
}

func (self *Mailer) logOutbox() {
	for status, messages := range self.Outbox {
		self.Logger.Printf("%s: %d", status, len(messages))
	}
}
func (self *Mailer) loop(stop <-chan struct{}) {
	everySecond := time.NewTicker(time.Second)
	defer everySecond.Stop()
	for {
		select {
		case <-stop:
			return
		case <-everySecond.C:
			self.catchUp()
			self.sendEmails()
		}
	}
}

func (self *Mailer) catchUp() {
	commands, err := self.App.Commands.After(self.Version)
	if err != nil {
		self.Logger.Printf("failed to fetch commands: %v", err)
		return
	}
	for command := range commands {
		self.HandleCommand(command.Message)
		self.Version = command.ID
	}
	if len(self.Outbox[StatusQueued]) > 0 {
		self.Logger.Printf("new messages in outbox")
		self.logOutbox()
	}
}

func (self *Mailer) deliver(id string, receipt EmailReceipt) {
	cmd := &SetEmailDeliveryStatus{
		InternalID: id,
		Status:     StatusDelivered,
	}
	if receipt != nil {
		cmd.Message = receipt.ExternalMessageID()
	}
	if err := self.App.HandleCommand(cmd); err != nil {
		self.Logger.Printf("failed to set email status: %v", err)
		return
	}
}

func (self *Mailer) fail(id string, err error) {
	if err := self.App.HandleCommand(&SetEmailDeliveryStatus{
		InternalID: id,
		Status:     StatusFailed,
		Message:    err.Error(),
	}); err != nil {
		self.Logger.Printf("failed to set email status: %v", err)
		return
	}
	self.Logger.Printf("email %s failed: %v", id, err)
}

func (self *Mailer) sendEmails() {
	for _, email := range self.Outbox[StatusQueued] {
		if receipt, err := self.Sender.SendEmail(email); err != nil {
			self.Logger.Printf("failed to send email: %v", err)
			if email.Retries < 3 {
				self.Logger.Printf("email %s has %d retries left", email.InternalID, 3-email.Retries)
				email.Retries++
				continue
			} else {
				self.fail(email.InternalID, fmt.Errorf("retries exhausted: %w", err))
				continue
			}
		} else {
			self.deliver(email.InternalID, receipt)
		}
	}
}
