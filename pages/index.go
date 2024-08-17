package pages

import (
	"fmt"
	"net/url"
	"time"

	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"

	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

func TimeLabel(t time.Time) g.Node {
	return Time(
		DateTime(t.Format(time.RFC3339)),
		g.Text(t.Format("2006-01-02 15:04:05")),
	)
}

type Submission struct {
	ItemID         string
	ImageURL       *string
	Submitter      string
	SubmittedAt    time.Time
	Url            string
	Title          string
	GeneratedTitle string
	VoteCount      int
	CommentCount   int
	CanVote        bool
	Comments       []Comment
}

func (s *Submission) Byline() string {
	if s.GeneratedTitle == "" {
		return s.Url
	}

	return fmt.Sprintf("%s - %s", s.Url, s.GeneratedTitle)
}

type Comment interface {
	CommentAuthor() string
	CommentContent() string
	WrittenAt() time.Time
	CommentableID() string
}

type WithChildren interface {
	AllChildren() []interface{}
}

type User struct {
	Username string
}

type PageData struct {
	CurrentUser *User
	FormState   *FormState
	BackTo      *url.URL
}

func (p *PageData) Username() *string {
	if p.CurrentUser != nil {
		return &p.CurrentUser.Username
	}

	return nil
}

func IndexPage(path string, submissions []*Submission, context *PageData) g.Node {
	return Page("The Orange Website", path, SubmissionList(submissions), context)
}

func SubmissionList(submissions []*Submission) g.Node {
	counter := 0
	return Container(
		Class("flex flex-col space-y-2"),
		g.Group(g.Map(submissions, func(s *Submission) g.Node {
			counter += 1
			return Div(
				Class("flex flex-row space-x-2"),
				Data("item-id", s.ItemID),
				Div(Class("prose w-4 text-center"), g.Textf("%d.", counter)),
				g.Iff(s.ImageURL != nil, func() g.Node {
					return Div(Class(fmt.Sprintf("min-w-8 w-8 h-8 bg-[url('%s')] bg-contain bg-center bg-no-repeat", *s.ImageURL)))
				}),
				g.Iff(s.ImageURL == nil, func() g.Node {
					return Div(Class("min-w-8 w-8 h-8"))
				}),
				Div(
					P(
						A(Href(s.Url), g.Text(s.Title)),
						Span(Class("text-sm ml-1 break-words text-gray-400"),
							g.Textf("(%s)", s.Byline())),
					),
					Div(
						Class("prose text-xs"),
						g.Textf("%d points by %s | ", s.VoteCount, s.Submitter),
						g.If(s.CanVote, UpvoteButton(s.ItemID)),
						TimeLabel(s.SubmittedAt),
						g.Text(" | "),
						A(Href("/item?id="+s.ItemID), g.Textf("%d comments", s.CommentCount)),
					),
				))
		})))
}

func Page(title, path string, body g.Node, context *PageData) g.Node {
	// HTML5 boilerplate document
	return c.HTML5(c.HTML5Props{
		Title:    title,
		Language: "en",
		Head: []g.Node{
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
			Script(Src("https://unpkg.com/htmx.org@2.0.1")),
			Script(Src("https://unpkg.com/htmx-ext-sse@2.2.1/sse.js")),
		},
		Body: []g.Node{
			Class("m-0 flex min-h-screen flex-col"),
			Navbar(path, context),
			NotificationBar(),
			Main(Class("flex-auto"), body),
			PageFooter(),
		},
	})
}

type PageLink struct {
	Path string
	Name string
}

func NotificationBar() g.Node {
	return Div(
		hx.Ext("sse"),
		g.Attr("sse-connect", "/notify"),
		g.Attr("sse-swap", "notify"),
		Class("min-h-8 bg-amber-200 mb-4 p-2 text-sm text-center font-mono"),
		g.Text("Notifications will appear here"),
	)
}

func Navbar(currentPath string, context *PageData) g.Node {
	return Nav(Class("bg-orange-500"),
		Container(
			Div(Class("flex relative items-center min-h-16"),
				Span(Class("text-white font-bold p-2"), g.Text("The Orange Website")),
				NavbarLink("/", "New", currentPath == "/"),
				NavbarLink("/submit", "Submit", currentPath == "/submit"),

				g.If(context.CurrentUser == nil, NavbarLink("/login", "Log in", currentPath == "/login")),
				g.Iff(context.CurrentUser != nil,
					func() g.Node {
						return g.Group([]g.Node{
							NavbarLink("/me", context.CurrentUser.Username, currentPath == "/me"),
							NavbarLink("/logout", "Log out", false),
						})
					}),
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
	return Div(Class("sm:max-w-16 md:max-w-7xl mx-auto px-2 sm:px-6 lg:px-8"), g.Group(children))
}

func Prose(children ...g.Node) g.Node {
	return Div(Class("prose"), g.Group(children))
}

func PageFooter() g.Node {
	return Footer(Class("block"),
		P(Class("text-center text-sm text-gray-400"),
			ID("rendered-at"),
			// We can use string interpolation directly, like fmt.Sprintf.
			g.Textf("Rendered %v. ", time.Now().Format(time.RFC3339)),
		),
	)
}
