package pages

import (
	"time"

	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"

	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

func TimeLabel(t time.Time) g.Node {
	return Time(DateTime(t.Format(time.RFC3339)),
		g.Text(t.Format("2006-01-02 15:04:05")),
	)
}

type Submission struct {
	ItemID      string
	ImageURL    *string
	Submitter   string
	SubmittedAt time.Time
	Url         string
	Title       string
	VoteCount   int
	CanVote     bool
}

type User struct {
	Username string
}

type PageData struct {
	CurrentUser *User
	FormErrors  map[string]string
}

func (p *PageData) Username() *string {
	if p.CurrentUser != nil {
		return &p.CurrentUser.Username
	}

	return nil
}

func (p *PageData) HasFormErrors() bool {
	return len(p.FormErrors) > 0
}

func (p *PageData) AddFormError(field, message string) {
	if p.FormErrors == nil {
		p.FormErrors = map[string]string{}
	}

	p.FormErrors[field] = message
}

func IndexPage(path string, submissions []*Submission, context *PageData) g.Node {
	return Page("The Orange Website", path, SubmissionList(submissions), context)
}

func SubmissionList(submissions []*Submission) g.Node {
	counter := 0
	return Container(
		hx.Get("/"),
		hx.Trigger("load delay:1s"),
		hx.Swap("outerHTML"),
		Class("flex flex-col space-y-2"),
		g.Group(g.Map(submissions, func(s *Submission) g.Node {
			counter += 1
			return Div(Class("flex flex-row space-x-2"),
				Div(Class("prose w-4 text-center"), g.Textf("%d.", counter)),
				g.Iff(s.ImageURL != nil, func() g.Node {
					return Img(Src(*s.ImageURL), Alt(s.Title), Class("w-8 h-8"))
				}),
				g.Iff(s.ImageURL == nil, func() g.Node {
					return Div(Class("w-8 h-8"))
				}),
				Div(
					P(Class("prose"),
						A(Href(s.Url), g.Text(s.Title)),
						Span(Class("text-sm ml-1 text-gray-400"),
							g.Textf("(%s)", s.Url))),
					Div(Class("prose text-xs"),
						g.Textf("%d points by %s | ", s.VoteCount, s.Submitter),
						g.If(s.CanVote, UpvoteButton(s.ItemID)),
						TimeLabel(s.SubmittedAt)),
				))
		})))
}

func Page(title, path string, body g.Node, context *PageData) g.Node {
	// HTML5 boilerplate document
	return c.HTML5(c.HTML5Props{
		Title:    title,
		Language: "en",
		Head: []g.Node{
			Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
			Script(Src("https://unpkg.com/htmx.org@2.0.1")),
		},
		Body: []g.Node{
			Class("flex min-h-screen flex-col"),
			Navbar(path, context, []PageLink{{}}),
			Main(Class("flex-auto"), body),
			PageFooter(),
		},
	})
}

type PageLink struct {
	Path string
	Name string
}

func Navbar(currentPath string, context *PageData, links []PageLink) g.Node {
	return Nav(Class("bg-orange-500 mb-4"),
		Container(
			Div(Class("flex relative items-center space-x-4 h-16"),
				Span(Class("text-white font-bold"), g.Text("The Orange Website")),
				NavbarLink("/", "New", currentPath == "/"),
				NavbarLink("/submit", "Submit", currentPath == "/submit"),

				// We can Map custom slices to Nodes
				g.Group(g.Map(links, func(pl PageLink) g.Node {
					return NavbarLink(pl.Path, pl.Name, currentPath == pl.Path)
				})),
				Div(Class("absolute right-0"),
					g.If(context.CurrentUser == nil, NavbarLink("/login", "Log in", currentPath == "/login")),
					g.Iff(context.CurrentUser != nil,
						func() g.Node {
							return g.Group([]g.Node{Span(Class("prose"), g.Text(context.CurrentUser.Username)), NavbarLink("/logout", "Log out", false)})
						}),
				),
			),
		),
	)
}

// NavbarLink is a link in the Navbar.
func NavbarLink(path, text string, active bool) g.Node {
	return A(Href(path), g.Text(text),
		// Apply CSS classes conditionally
		c.Classes{
			"px-5 py-2 rounded-md text-sm font-medium focus:outline-none focus:text-white focus:bg-orange-700": true,
			"text-white bg-orange-700":                        active,
			"text-white hover:text-white hover:bg-orange-700": !active,
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
