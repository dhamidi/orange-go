package main

type AuthMagicLinkLogin struct {
	Email string
}

func (cmd *AuthMagicLinkLogin) CommandName() string {
	return "AuthMagicLinkLogin"
}

func init() {
	DefaultCommandRegistry.Register("AuthMagicLinkLogin", func() Command { return new(AuthMagicLinkLogin) })
}

func (self *Auth) handleAuthMagicLinkLogin(cmd *AuthMagicLinkLogin) error {
	panic("implement me")
}
