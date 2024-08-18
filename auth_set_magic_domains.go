package main

type SetMagicDomains struct {
	Domains []string
}

func (cmd *SetMagicDomains) CommandName() string {
	return "SetMagicDomains"
}

func init() {
	DefaultCommandRegistry.Register("SetMagicDomains", func() Command { return new(SetMagicDomains) })
}

func (self *Auth) handleSetMagicDomains(cmd *SetMagicDomains) error {
	self.state.SetMagicDomains(cmd.Domains)
	return nil
}
