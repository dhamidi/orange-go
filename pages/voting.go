package pages

import (
	g "github.com/maragudk/gomponents"

	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
)

func VotedIcon() g.Node {
	return Span(Class("font-mono"), g.Text("[__]"))
}
func UpvoteButton(itemID string) g.Node {
	return Form(
		Class("inline"),
		hx.Boost("true"),
		hx.Target("this"),
		hx.PushURL("false"),
		Action("/upvote"),
		Method("POST"),
		Input(
			Type("hidden"),
			Name("itemID"),
			Value(itemID),
		),
		Button(
			Class("inline font-mono"),
			Type("submit"),
			g.Text("[+1]"),
		),
	)
}
