package pages

import (
	"fmt"
	"net/url"
	"runtime"
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
	Index          uint64
	ItemID         string
	ImageURL       *string
	Submitter      string
	SubmittedAt    time.Time
	Url            string
	Title          string
	GeneratedTitle string
	Hidden         bool
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
	IsHidden() bool
	CommentAuthor() string
	CommentContent() string
	WrittenAt() time.Time
	CommentableID() string
	CommentID() string
}

type WithChildren interface {
	AllChildren() []interface{}
}

type User struct {
	Username string
}

type PageData struct {
	Stylesheet  string
	CurrentUser *User
	IsAdmin     bool
	FormState   *FormState
	BackTo      *url.URL
	LoadMore    *url.URL
	MainOnly    bool
	OpenGraph   OpenGraph
}

type OpenGraph struct {
	Title       string
	Description string
	Type        string
	Url         string
	ImageURL    string
}

func (og *OpenGraph) Render() g.Node {
	return g.Group([]g.Node{
		g.If(og.Title != "", Meta(g.Attr("property", "og:title"), Content(og.Title))),
		g.If(og.Description != "", Meta(g.Attr("property", "og:description"), Content(og.Description))),
		g.If(og.Type != "", Meta(g.Attr("property", "og:type"), Content(og.Type))),
		g.If(og.Url != "", Meta(g.Attr("property", "og:url"), Content(og.Url))),
		g.If(og.ImageURL != "", Meta(g.Attr("property", "og:image"), Content(og.ImageURL))),
	})
}

func (p *PageData) Navlinks() []*PageLink {
	result := []*PageLink{}
	result = append(result, &PageLink{Path: "/", Name: "Top"})
	result = append(result, &PageLink{Path: "/submit", Name: "Submit"})
	if p.CurrentUser == nil {
		result = append(result, &PageLink{Path: "/login", Name: "Log in"})
	} else {
		result = append(result,
			&PageLink{Path: "/me", Name: p.CurrentUser.Username},
			&PageLink{Path: "/logout", Name: "Log out"},
		)
	}

	return result
}

func (p *PageData) Username() *string {
	if p.CurrentUser != nil {
		return &p.CurrentUser.Username
	}

	return nil
}

func IndexPage(path string, submissions []*Submission, context *PageData) g.Node {
	return Page("The Orange Website", path, SubmissionList(submissions, context.LoadMore, context.IsAdmin), context)
}

func SubmissionList(submissions []*Submission, loadMore *url.URL, isAdmin bool) g.Node {
	return Container(
		Class("flex flex-col space-y-2"),
		g.Group(g.Map(submissions, func(s *Submission) g.Node {
			return Div(
				Class("flex flex-row space-x-2"),
				Data("item-id", s.ItemID),
				Div(
					P(
						A(Href(s.Url), g.Textf("%d. %s", s.Index, s.Title)),
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
						g.If(
							isAdmin,
							g.Group([]g.Node{
								g.Text(" | "),
								g.If(s.Hidden, UnhideSubmissionButton(s.ItemID)),
								g.If(!s.Hidden, HideSubmissionButton(s.ItemID)),
							}),
						),
					),
				))
		})),
		g.Iff(loadMore != nil, func() g.Node {
			return Div(Class("mt-4"),
				ButtonLink("More", loadMore.String(), hx.PushURL("true"), hx.Target("main"), hx.Get(loadMore.String())),
			)
		}),
	)
}

func Page(title, path string, body g.Node, context *PageData) g.Node {
	if context.MainOnly {
		return body
	}
	// HTML5 boilerplate document
	return c.HTML5(c.HTML5Props{
		Title:    title,
		Language: "en",
		Head: []g.Node{
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			Script(Src("/s/htmx.min.a651db4.js")),
			Script(Src("/s/htmx-sse.713ef8d.js")),
			Script(Src("/s/alpine-3.14.1.min.cd31b85.js"), Defer()),
			Link(Href("/s/"+context.Stylesheet), Rel("stylesheet")),
			context.OpenGraph.Render(),
		},
		Body: []g.Node{
			Class("m-0 font-mono flex min-h-screen flex-col"),
			Navbar(path, context),
			// NotificationBar() currently disabled because I couldn't find a way to close EventSources after navigation.
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
	bar := Div(
		hx.Ext("sse"),
		g.Attr("sse-connect", "/notify"),
		g.Attr("sse-swap", "notify"),
		Class("min-h-8 bg-amber-200 mb-4 p-2 text-sm text-center font-mono"),
		g.Text("Notifications will appear here"),
	)
	return bar
}

func Navbar(currentPath string, context *PageData) g.Node {
	return Nav(
		Class("pb-2"),
		NavbarLarge(currentPath, context),
		NavbarSmall(currentPath, context),
	)
}

func NavbarSmall(currentPath string, context *PageData) g.Node {
	return Div(Class("bg-orange-500 flex flex-col h-full w-full lg:hidden"),
		g.Attr("x-data", "{ navbarOpen: false }"),
		Div(Class("w-full flex flex-row justify-between"),
			Span(
				Class("inline-block text-left w-32 h-16"),
				Img(Class("inline-block h-16"), Src("/s/android-chrome-192x192.png")),
			),
			Div(Class("flex flex-col justify-center cursor-pointer text-white px-2 text-right w-32"),
				g.Attr("@click", "navbarOpen = !navbarOpen"),
				g.Text("[MENU]"),
			),
		),
		Div(
			g.Attr("x-show", "navbarOpen"),
			Class("h-screen"),
			g.Group(g.Map(context.Navlinks(), func(link *PageLink) g.Node {
				return NavbarLinkSmall(link.Path, link.Name, currentPath == link.Path)
			})),
		),
	)
}

func NavbarLarge(currentPath string, context *PageData) g.Node {
	return Div(Class("bg-orange-500 hidden lg:block"),
		Container(
			Div(Class("flex relative items-center min-h-16"),
				A(Href("/"), Class("text-white font-bold p-2"), g.Text("The Orange Website")),
				g.Group(g.Map(context.Navlinks(), func(link *PageLink) g.Node {
					return NavbarLinkLarge(link.Path, link.Name, currentPath == link.Path)
				})),
			),
		),
	)
}

func NavbarLinkLarge(path, text string, active bool) g.Node {
	return A(Href(path), g.Text(text),
		c.Classes{
			"px-5 py-2 text-sm font-medium focus:outline-none focus:text-white focus:bg-orange-700": true,
			"text-white bg-orange-700":                        active,
			"text-white hover:text-white hover:bg-orange-700": !active,
		},
	)
}

func NavbarLinkSmall(path, text string, active bool) g.Node {
	return A(Href(path), g.Text(text),
		c.Classes{
			"px-5 py-2 text-xl font-medium focus:outline-none focus:text-white focus:bg-orange-700 block": true,
			"text-orange-500 bg-white":                        active,
			"text-white hover:text-white hover:bg-orange-700": !active,
		},
	)
}

func Container(children ...g.Node) g.Node {
	return Div(Class("sm:max-w-16 md:max-w-7xl mx-auto px-2 sm:px-6 lg:px-8"), g.Group(children))
}

func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n := n / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func memoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("MEM: %v", humanBytes(m.Alloc))
}

func PageFooter() g.Node {
	return Footer(Class("block"),
		P(Class("text-center font-mono text-sm text-gray-400"),
			ID("rendered-at"),
			g.Textf("T %v ", time.Now().Format(time.RFC3339)),
			g.Textf(" | %s", memoryUsage()),
		),
	)
}
