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

func Test_LinkVerifiedEmailToUser_UpdatesTheUserObject(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("admin", "admin"))
	scenario.must(scenario.linkVerifiedEmailToUser("admin", "admin@example.com"))
	user, err := scenario.findUserByEmail("admin@example.com")
	if err != nil {
		t.Fatalf("failed to find user by email: %s", err)
	}

	if user.VerifiedEmail != "admin@example.com" {
		t.Fatalf("expected verified email to be %q, got %q", "admin@example.com", user.VerifiedEmail)
	}
}

func Test_FindUserByEmail_ReturnsError_WhenEmailIsNotLinked(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("admin", "admin"))
	_, err := scenario.findUserByEmail("admin@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected %s, got %s", ErrUserNotFound, err)
	}
}

func Test_LinkVerifiedEmailToUser_ReturnsError_WhenUserDoesNotExist(t *testing.T) {
	scenario := setup(t)
	scenario.mustFailWith(scenario.linkVerifiedEmailToUser("admin", "admin@example.com"), ErrUserNotFound)
}

func Test_RequestMagicLinkLogin_ReturnsError_WhenUserDoesNotExist(t *testing.T) {
	scenario := setup(t)
	scenario.mustFailWith(scenario.requestMagicLinkLogin("test@gmail.com", "magic-string"), ErrUserNotFound)
}

func Test_RequestMagicLinkLogin_UpdatesMagic(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("admin", "admin"))
	scenario.must(scenario.linkVerifiedEmailToUser("admin", "admin@example.com"))
	scenario.must(scenario.requestMagicLinkLogin("admin@example.com", "magic-string"))

	user, err := scenario.findUserByEmail("admin@example.com")
	if err != nil {
		t.Fatalf("failed to find user by email: %s", err)
	}

	if user.Magic != "magic-string" {
		t.Fatalf("expected magic to be %q, got %q", "magic-string", user.Magic)
	}
}

func Test_RequestMagicLinkLogin_UpdatesMagic_IfUserAlreadyHasMagic(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("admin", "admin"))
	scenario.must(scenario.linkVerifiedEmailToUser("admin", "admin@example.com"))
	scenario.must(scenario.requestMagicLinkLogin("admin@example.com", "magic-string"))
	scenario.must(scenario.requestMagicLinkLogin("admin@example.com", "super-magic-string"))

	user, err := scenario.findUserByEmail("admin@example.com")
	if err != nil {
		t.Fatalf("failed to find user by email: %s", err)
	}

	if user.Magic != "super-magic-string" {
		t.Fatalf("expected magic to be %q, got %q", "super-magic-string", user.Magic)
	}
}

func Test_LogInWithMagic_Succeeds_WhenUserExistsAndMagicMatches(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.signup("admin", "admin"))
	scenario.must(scenario.linkVerifiedEmailToUser("admin", "admin@example.com"))
	scenario.must(scenario.requestMagicLinkLogin("admin@example.com", "magic-string"))

	user, err := scenario.findUserByEmail("admin@example.com")
	if err != nil {
		t.Fatalf("failed to find user by email: %s", err)
	}

	scenario.must(scenario.loginWithMagic(user.Magic))
}

func Test_LogInWithMagic_ReturnsError_WhenMagicIsForbidden(t *testing.T) {
	scenario := setup(t)
	scenario.mustFailWith(scenario.loginWithMagic("magic-string"), ErrUserNotFound)
}

func Test_LogInWithMagic_CreatesNewUser_WhenDomainIsMagic(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.setMagicDomains("bolt.eu"))
	scenario.must(scenario.requestMagicLinkLogin("dario.hamidi@bolt.eu", "magic-string"))
	scenario.must(scenario.loginWithMagic("magic-string"))

	user, err := scenario.findUserByEmail("dario.hamidi@bolt.eu")
	if err != nil {
		t.Fatalf("failed to find user by email: %s", err)
	}

	if act, exp := user.Username, "dario.hamidi"; act != exp {
		t.Fatalf("expected username to be %q, got %q", exp, act)
	}
}
