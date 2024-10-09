package main

import (
	"testing"
	"time"
)

func TestContentSubscriptions_Disabling(t *testing.T) {
	settings := NewDefaultSubscriptionSettings("admin", time.Now())
	if len(settings.EnabledFor) != 0 {
		t.Fatalf("expected no settings to be enabled, got %v", settings.EnabledFor)
	}

	settings.EnableScope(SUBSCRIPTION_SCOPE_REPLIES)
	if len(settings.EnabledFor) != 1 {
		t.Fatalf("expected one setting to be enabled, got %v", settings.EnabledFor)
	}

	settings.EnableScope(SUBSCRIPTION_SCOPE_SUBMISSIONS)
	if len(settings.EnabledFor) != 2 {
		t.Fatalf("expected two settings to be enabled, got %v", settings.EnabledFor)
	}

	if len(settings.DisabledFor) != 0 {
		t.Fatalf("expected no settings to be disabled, got %v", settings.DisabledFor)
	}
}
