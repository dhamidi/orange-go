package main

type SetAdminUsers struct {
	Users []string
}

func (cmd *SetAdminUsers) CommandName() string {
	return "SetAdminUsers"
}

func init() {
	DefaultCommandRegistry.Register("SetAdminUsers", func() Command { return &SetAdminUsers{} })
}

func (self *Auth) handleSetAdminUsers(cmd *SetAdminUsers) error {
	return self.state.SetAdminUsers(cmd.Users)
}
