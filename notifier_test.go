package main

import (
	"testing"
)

func Test_Notifier_does_notify_about_old_entries(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("test-user", "password"))
	scenario.must(scenario.linkVerifiedEmailToUser("test-user", "test-user@gmail.local"))
	scenario.must(scenario.signup("dontnotify", "password"))
	scenario.must(scenario.linkVerifiedEmailToUser("dontnotify", "dontnotify@gmail.local"))
	scenario.must(scenario.postLink("https://en.wikipedia.org", "Wiki"))
	scenario.Viewer = "dontnotify"
	notifier := scenario.Notifier()
	notifier.catchUp()
	notifier.notify()

	scenario.must(scenario.subscribeTo("test-user", SUBSCRIPTION_SCOPE_SUBMISSIONS))
	scenario.must(scenario.postLink("https://example.com", "A new entry"))
	submissions := scenario.frontpage()
	if len(submissions) == 0 {
		t.Fatalf("Frontpage unexpectedly empty")
	}
	firstPostID := submissions[0].ItemID
	scenario.must(scenario.commentOn(firstPostID, "does not trigger a notification"))
	notifier.catchUp()

	if len(notifier.ToNotify) == 0 {
		t.Fatalf("Expected a notification to be scheduled, but there is none")
	}

	notifier.notify()
	scenario.DumpLog()

	notificationQueuedFor := func(recipientEmail string) func(cmd *PersistedCommand) bool {
		return func(cmd *PersistedCommand) bool {
			if cmd.Message.CommandName() != "QueueEmail" {
				return false
			}
			queueEmail := cmd.Message.(*QueueEmail)
			return queueEmail.TemplateName == "content-notification" && queueEmail.Recipients == recipientEmail
		}
	}
	if !scenario.LogContains(notificationQueuedFor("test-user@gmail.local")) {
		t.Fatalf("Expected an email to be queued, found none")
	}
	if scenario.LogContains(notificationQueuedFor("dontnotify@gmail.local")) {
		t.Fatalf("Expected no email to be queued for dontnotify@gmail.local")
	}
}
