package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

// Notifier is a background process that notifies users of new
// submissions and comments.
//
// When a PostLink or PostComment entry is found in the log,
// a query is sent to the content module to ask which user should
// receive notifications based on their subscription settings.
type Notifier struct {
	Logger   *log.Logger
	App      *App
	Commands CommandLog
	ToNotify ScheduledNotificationSet
	Version  int
	BaseURL  *url.URL
}

type ScheduledNotificationSet map[string]*ScheduledNotification

func (set ScheduledNotificationSet) Add(n *ScheduledNotification) ScheduledNotificationSet {
	set[n.ID()] = n
	return set
}

func (set ScheduledNotificationSet) Remove(n *ScheduledNotification) ScheduledNotificationSet {
	delete(set, n.ID())
	return set
}

type ScheduledNotification struct {
	About     string
	Recipient string
	Event     string
}

func (n *ScheduledNotification) Entity() string {
	switch n.Event {
	case "PostComment":
		return "comment"
	case "PostLink":
		return "submission"
	default:
		return n.Event
	}
}

func (n *ScheduledNotification) ID() string { return fmt.Sprintf("%s:%s", n.About, n.Recipient) }

func NewNotifier(app *App, commands CommandLog, logger *log.Logger, baseURL *url.URL) *Notifier {
	return &Notifier{
		Logger:   logger,
		App:      app,
		Commands: commands,
		ToNotify: make(ScheduledNotificationSet),
		Version:  0,
		BaseURL:  baseURL,
	}
}
func (n *Notifier) Start() func() {
	stop := make(chan struct{})

	n.catchUp()
	n.Logger.Printf("Notifier started at version %d", n.Version)
	go n.loop(stop)
	return func() { close(stop) }
}

func (n *Notifier) catchUp() {
	commands, err := n.Commands.After(n.Version)
	if err != nil {
		n.Logger.Printf("failed to fetch commands: %v", err)
		return
	}
	for command := range commands {
		n.HandleCommand(command.Message)
		n.Version = command.ID
	}
}

func (n *Notifier) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	tick := ticker.C
	for {
		select {
		case <-stop:
			return
		case <-tick:
			n.catchUp()
			n.showWork()
			n.notify()
		}
	}
}

func (n *Notifier) showWork() {
	if len(n.ToNotify) == 0 {
		return
	}

	for _, notification := range n.ToNotify {
		n.Logger.Printf("%q about %q(%q)", notification.Recipient, notification.Event, notification.About)
	}
}

func (n *Notifier) HandleCommand(cmd Command) {
	switch cmd := cmd.(type) {
	case *QueueEmail:
		n.removeScheduleNotificationFor(cmd)
	case *PostLink:
		n.addScheduledNotificationForSubmission(cmd)
	case *PostComment:
		n.addScheduledNotificationForComment(cmd)
	}
}

func (s *Notifier) notify() {
	scheduled := []*ScheduledNotification{}
	for _, n := range s.ToNotify {
		scheduled = append(scheduled, n)
	}
	for _, n := range scheduled {
		s.notifyAbout(n)
		s.ToNotify.Remove(n)
	}
}

func (n *Notifier) notifyAbout(notification *ScheduledNotification) {
	entity := notification.Entity()
	submission := (*Submission)(nil)
	submissionID := notification.About
	if entity == "comment" {
		submissionID = NewTreeID(notification.About).Root()
	}
	q := NewFindSubmission(submissionID)
	if err := n.App.HandleQuery(q); err != nil {
		n.Logger.Printf("notifyAbout(%q): failed to find submission(%q): %s", notification.Event, notification.About, err)
		return
	}
	submission = q.Submission
	title := submission.Title

	actionURL := n.BaseURL.JoinPath("item")
	id := actionURL.Query()
	id.Add("id", notification.About)
	actionURL.RawQuery = id.Encode()

	recipientEmail, found := n.findRecipientEmail(notification.Recipient)
	if !found {
		n.Logger.Printf("skipping %q, no verified email found", notification.Recipient)
		return
	}

	queueEmail := &QueueEmail{
		InternalID:   notification.ID(),
		Recipients:   recipientEmail,
		Subject:      "",
		TemplateName: "content-notification",
		TemplateData: map[string]any{
			"about":      notification.About,
			"trigger":    notification.Event,
			"name":       notification.Recipient,
			"title":      title,
			"action_url": actionURL.String(),
			"entity":     entity,
		},
	}
	n.App.HandleCommand(queueEmail)
}

func (n *Notifier) findRecipientEmail(username string) (string, bool) {
	q := NewFindUserByName(username)
	if err := n.App.HandleQuery(q); err != nil {
		n.Logger.Printf("findRecipientEmail: user %q not found: %s", username, err)
		return "", false
	}

	if q.User.VerifiedEmail == "" {
		return "", false
	}

	return q.User.VerifiedEmail, true
}

func (n *Notifier) removeScheduleNotificationFor(cmd *QueueEmail) {
	if cmd.TemplateName != "content-notification" {
		return
	}
	entityIDAndRecipient := strings.Split(cmd.InternalID, ":")
	about, recipient := entityIDAndRecipient[0], entityIDAndRecipient[1]
	trigger := "unknown"
	if t, found := cmd.TemplateData["trigger"]; found {
		if ts, ok := t.(string); ok {
			trigger = ts
		}
	}
	notification := &ScheduledNotification{
		About:     about,
		Event:     trigger,
		Recipient: recipient,
	}

	n.ToNotify.Remove(notification)
}
func (n *Notifier) addScheduledNotificationForSubmission(cmd *PostLink) {
	for _, recipient := range n.recipientsFor(cmd) {
		n.Logger.Printf("scheduling for %q", recipient)
		n.ToNotify.Add(&ScheduledNotification{
			About:     cmd.ItemID,
			Event:     "PostLink",
			Recipient: recipient,
		})
	}
}
func (n *Notifier) addScheduledNotificationForComment(cmd *PostComment) {
	for _, recipient := range n.recipientsFor(cmd) {
		n.Logger.Printf("scheduling for %q", recipient)
		n.ToNotify.Add(&ScheduledNotification{
			About:     cmd.ParentID.String(),
			Event:     "PostComment",
			Recipient: recipient,
		})
	}
}

func (n *Notifier) recipientsFor(cmd Command) []string {
	switch cmd := cmd.(type) {
	case *PostLink:
		q := NewFindSubscribersForNewSubmission()
		n.App.HandleQuery(q)
		return q.Result().([]string)
	case *PostComment:
		q := NewFindSubscribersForNewComment(cmd.ParentID.String())
		n.App.HandleQuery(q)
		return q.Result().([]string)
	default:
		return []string{}
	}
}
