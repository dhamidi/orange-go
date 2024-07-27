package main

import (
	"errors"
	"testing"
)

func Test_ASignedUpUser_CanLogIn_WithTheirPassword(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("test-user", "test-password"))
	hashedPassword, err := scenario.findPasswordHash("test-user", "test-password")
	if err != nil {
		t.Fatalf("failed to find password hash: %s", err)
	}
	scenario.must(scenario.login("test-user", hashedPassword))
}

func Test_ASignedUpUser_CannotLogIn_WithWrongPassword(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("test-user", "test-password"))
	_, err := scenario.findPasswordHash("test-user", "wrong-password")
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("expected ErrPasswordMismatch, got %s", err)
	}
}
