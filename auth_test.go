package main

import (
	"errors"
	"testing"
)

func Test_ASignedUpUser_CanLogIn_WithTheirPassword(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("test-user", "test-password"))
	_, err := scenario.findPasswordHash("test-user", "test-password")
	if err != nil {
		t.Fatalf("failed to find password hash: %s", err)
	}
	scenario.must(scenario.login("test-user", "test-password"))
}

func Test_ASignedUpUser_CannotLogIn_WithWrongPassword(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("test-user", "test-password"))
	_, err := scenario.findPasswordHash("test-user", "wrong-password")
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("expected ErrPasswordMismatch, got %s", err)
	}
}

func Test_ASignedUpUser_CannotSignUpAgain(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("admin", "admin"))
	scenario.mustFailWith(scenario.signup("admin", "admin"), ErrUserExists)
}

func Test_ASignedUpUser_MustRespectTheUsernamePolicy(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("a", "admin"))
	scenario.must(scenario.setUsernamePolicy(2, 100))
	scenario.mustFailWith(scenario.signup("b", "password"), ErrUsernameNotAllowed)
}

func Test_UsernamePolicy_allows_excluding_names(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.forbidUsername("guest", "admin"))
	scenario.mustFailWith(scenario.signup("admin", "safe-password"), ErrUsernameNotAllowed)
	scenario.mustFailWith(scenario.signup("guest", "safe-password"), ErrUsernameNotAllowed)
	scenario.must(scenario.signup("regular-user", "safe-password"))
}
