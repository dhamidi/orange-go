package pages

import (
	"time"

	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"

	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

type Submission struct {
	Submitter string
	Url       string
	Title     string
}

func IndexPage(path string, submissions []*Submission) g.Node {
	return Page("The Orange Website", path, SubmissionList(submissions))
}

func SubmissionList(submissions []*Submission) g.Node {
	return Div(
		hx.Get("/"),
		hx.Trigger("load delay:1s"),
		hx.Swap("outerHTML"),
		Class("flex flex-col space-y-2"),
		g.Group(g.Map(submissions, func(s *Submission) g.Node {
			return P(A(Href(s.Url), g.Text(s.Title)))
		})))
}

func Page(title, path string, body g.Node) g.Node {
	// HTML5 boilerplate document
	return c.HTML5(c.HTML5Props{
		Title:    title,
		Language: "en",
		Head: []g.Node{
			Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
			Script(Src("https://unpkg.com/htmx.org@2.0.1")),
		},
		Body: []g.Node{
			Navbar(path, []PageLink{}),
			Container(
				Prose(body),
				PageFooter(),
			),
		},
	})
}

type PageLink struct {
	Path string
	Name string
}

func Navbar(currentPath string, links []PageLink) g.Node {
	return Nav(Class("bg-orange-500 mb-4"),
		Container(
			Div(Class("flex items-center space-x-4 h-16"),
				Span(Class("text-white font-bold"), g.Text("The Orange Website")),
				NavbarLink("/", "New", currentPath == "/"),

				// We can Map custom slices to Nodes
				g.Group(g.Map(links, func(pl PageLink) g.Node {
					return NavbarLink(pl.Path, pl.Name, currentPath == pl.Path)
				})),
			),
		),
	)
}

// NavbarLink is a link in the Navbar.
func NavbarLink(path, text string, active bool) g.Node {
	return A(Href(path), g.Text(text),
		// Apply CSS classes conditionally
		c.Classes{
			"px-3 py-2 rounded-md text-sm font-medium focus:outline-none focus:text-white focus:bg-orange-700": true,
			"text-white bg-orange-700":                           active,
			"text-gray-300 hover:text-white hover:bg-orange-700": !active,
		},
	)
}

func Container(children ...g.Node) g.Node {
	return Div(Class("max-w-7xl mx-auto px-2 sm:px-6 lg:px-8"), g.Group(children))
}

func Prose(children ...g.Node) g.Node {
	return Div(Class("prose"), g.Group(children))
}

func PageFooter() g.Node {
	return Footer(Class("prose prose-sm prose-indigo"),
		P(
			// We can use string interpolation directly, like fmt.Sprintf.
			g.Textf("Rendered %v. ", time.Now().Format(time.RFC3339)),
		),
	)
}
