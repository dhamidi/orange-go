package pages

import (
	g "github.com/maragudk/gomponents"

	. "github.com/maragudk/gomponents/html"
)

type AccountDetails struct {
	Username                string
	Email                   string
	EmailVerified           bool
	SubscribedToSubmissions bool
	SubscribedToReplies     bool
}

func MePage(details *AccountDetails, context *PageData) g.Node {
	return Page("The Orange Website | Me", "/me", AccountDetail(details), context)
}

func AccountDetail(details *AccountDetails) g.Node {
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		Div(
			Class("sm:mx-auto sm:w-full sm:max-w-sm"),
			H2(Class("mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"),
				g.Textf("👋 Hey %s", details.Username))),
		Div(
			Class("mt-10 sm:mx-auto sm:w-full sm:max-w-sm"),
			g.If(details.Email != "", g.Group([]g.Node{
				P(g.Textf("Your email is: %s", details.Email)),
				SubscriptionSettings(details),
			})),
			g.If(details.Email == "", P(g.Text("Once you link your email address, it'll be visible here."))),
		),
	)
}

func SubscriptionSettings(details *AccountDetails) g.Node {
	return Form(
		Class("flex flex-col mt-2"),
		Method("POST"),
		Action("/me"),
		H1(Class("font-bold text-xl"), g.Text("Get an email for")),
		P(
			Input(
				Class("mr-2"),
				Type("checkbox"),
				ID("subscribe_to_submissions"),
				Name("subscribe_to_submissions"),
				g.If(details.SubscribedToSubmissions, Checked()),
			),
			Label(For("subscribe_to_submissions"), g.Text("new submissions")),
		),
		P(
			Input(
				Class("mr-2"),
				Type("checkbox"),
				ID("subscribe_to_replies"),
				Name("subscribe_to_replies"),
				g.If(details.SubscribedToReplies, Checked()),
			),
			Label(For("subscribe_to_replies"), g.Text("new replies")),
		),
		SubmitButton("Save"),
	)
}
