package pages

import (
	g "github.com/maragudk/gomponents"

	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
)

func HiddenIcon() g.Node {
	return Span(Class("font-mono text-yellow-500"), g.Text("[hidden]"))
}
func HideSubmissionButton(itemID string, isHidden bool) g.Node {
	if isHidden {
		return HiddenIcon()
	}
	return Form(
		Class("inline"),
		hx.Boost("true"),
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
