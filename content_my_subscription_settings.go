package main

type SubscriptionSettingsForUser struct {
	Username string

	Settings *SubscriptionSettings
}

func (q *SubscriptionSettingsForUser) QueryName() string { return "MySubscriptionSettings" }
func (q *SubscriptionSettingsForUser) Result() any       { return q.Settings }

func NewMySubscriptionSettings(username string) *SubscriptionSettingsForUser {
	return &SubscriptionSettingsForUser{Username: username}
}

func (self *Content) findSubscriptionSettings(q *SubscriptionSettingsForUser) error {
	settings, err := self.state.GetSubscriptionSettings(q.Username)
	if err != nil {
		return err
	}
	q.Settings = settings
	return nil
}
