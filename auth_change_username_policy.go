package main

type ChangeUsernamePolicy struct {
	MinLength int
	MaxLength int
	Blacklist []string
}

func (cmd *ChangeUsernamePolicy) CommandName() string { return "ChangeUsernamePolicy" }

func init() {
	DefaultSerializer.Register("ChangeUsernamePolicy", func() Command { return new(ChangeUsernamePolicy) })
}

func (self *Auth) handleChangeUsernamePolicy(cmd *ChangeUsernamePolicy) error {
	return self.state.PutPolicy(&UsernamePolicy{
		MinLength: cmd.MinLength,
		MaxLength: cmd.MaxLength,
		Blacklist: cmd.Blacklist,
	})
}
