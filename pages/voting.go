package pages

import (
	g "github.com/maragudk/gomponents"

	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
)

func VotedIcon() g.Node {
	return Span(Class("inline mx-1 font-mono"), g.Text("[Voted +1]"))
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
			Class("inline font-mono mx-1"),
			Type("submit"),
			g.Text("[Upvote]"),
		),
	)
}
