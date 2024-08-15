package pages

import (
	g "github.com/maragudk/gomponents"

	. "github.com/maragudk/gomponents/html"
)

func ForbiddenMagicPage(path string) g.Node {
	return Page(
		"The Orange Website",
		path,
		Div(
			Class("flex flex-row justify-center items-center py-12"),
			P(Class("text-xl font-semibold"), g.Textf("Forbidden Magic 🧙")),
		),
		&PageData{},
	)
}
