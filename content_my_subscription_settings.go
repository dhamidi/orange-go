package main

type MySubscriptionSettings struct {
	Username string

	Settings *SubscriptionSettings
}

func (q *MySubscriptionSettings) QueryName() string { return "MySubscriptionSettings" }
func (q *MySubscriptionSettings) Result() any       { return q.Settings }

func NewMySubscriptionSettings(username string) *MySubscriptionSettings {
	return &MySubscriptionSettings{Username: username}
}

func (self *Content) findSubscriptionSettings(q *MySubscriptionSettings) error {
	settings, err := self.state.GetSubscriptionSettings(q.Username)
	if err != nil {
		return err
	}
	q.Settings = settings
	return nil
}
