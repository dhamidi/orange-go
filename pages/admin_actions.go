package pages

import (
	g "github.com/maragudk/gomponents"

	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
)

func UnhideSubmissionButton(itemID string) g.Node {
	return Form(
		Class("inline"),
		hx.Boost("true"),
		hx.PushURL("false"),
		hx.Target("this"),
		Action("/admin/a/unhide-submission"),
		Method("POST"),
		Input(
			Type("hidden"),
			Name("itemID"),
			Value(itemID),
		),
		Button(
			Class("inline font-mono font-bold text-red-500"),
			Type("submit"),
			g.Text("[unhide]"),
		),
	)
}
func HideSubmissionButton(itemID string) g.Node {
	return Form(
		Class("inline"),
		hx.Boost("true"),
		hx.PushURL("false"),
		hx.Target("this"),
		Action("/admin/a/hide-submission"),
		Method("POST"),
		Input(
			Type("hidden"),
			Name("itemID"),
			Value(itemID),
		),
		Button(
			Class("inline font-mono font-bold text-red-500"),
			Type("submit"),
			g.Text("[hide]"),
		),
	)
}

func UnhideCommentButton(itemID string) g.Node {
	return Form(
		Class("inline"),
		hx.Boost("true"),
		hx.PushURL("false"),
		hx.Target("this"),
		Action("/admin/a/unhide-comment"),
		Method("POST"),
		Input(
			Type("hidden"),
			Name("itemID"),
			Value(itemID),
		),
		Button(
			Class("inline font-mono font-bold text-red-500"),
			Type("submit"),
			g.Text("[unhide]"),
		),
	)
}
func HideCommentButton(itemID string) g.Node {
	return Form(
		Class("inline"),
		hx.Boost("true"),
		hx.PushURL("false"),
		hx.Target("this"),
		Action("/admin/a/hide-comment"),
		Method("POST"),
		Input(
			Type("hidden"),
			Name("itemID"),
			Value(itemID),
		),
		Button(
			Class("inline font-mono font-bold text-red-500"),
			Type("submit"),
			g.Text("[hide]"),
		),
	)
}
