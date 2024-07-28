package pages

import (
	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"

	. "github.com/maragudk/gomponents/html"
)

func ItemPage(path string, submission *Submission, context *PageData) g.Node {
	return Page("The Orange Website", path, SubmissionDetail(submission), context)
}

func SubmissionDetail(s *Submission) g.Node {
	return Container(
		Class("flex flex-col space-y-2"),
		Div(
			Class("mb-4"),
			P(Class("prose"),
				A(Href(s.Url), g.Text(s.Title)),
				Span(Class("text-sm ml-1 text-gray-400"),
					g.Textf("(%s)", s.Url))),
			Div(Class("prose text-xs"),
				g.Textf("%d points by %s | ", s.VoteCount, s.Submitter),
				g.If(s.CanVote, UpvoteButton(s.ItemID)),
				TimeLabel(s.SubmittedAt),
				g.Textf(" | %d comments", s.CommentCount)),
			Div(
				Class("my-2"),
				CommentForm(s.ItemID, NewFormState())),
		),
		g.Group(g.Map(s.Comments, func(c Comment) g.Node {
			return CommentWithChildren(c)
		})),
	)
}

func CommentLink(itemID string) g.Node {
	return A(
		hx.Get("/comment?item_id="+itemID),
		hx.Swap("outerHTML"),
		Href("/item?id="+itemID),
		Span(Class("text-xs font-mono"), g.Text("[reply]")),
	)
}

func CommentForm(itemID string, state *FormState) g.Node {
	return Form(
		hx.Post("/comment"),
		hx.Swap("this"),
		Action("/comment"), Method("POST"),
		Div(
			Class("flex flex-row space-around"),
			Input(Type("hidden"), Name("item_id"), Value(itemID)),
			InlineText("content", state),
			InlineSubmitButton("Reply"),
		),
	)
}

func CommentWithChildren(c Comment) g.Node {
	comment := CommentBlock(c)
	if hasChildren, ok := c.(WithChildren); ok {
		children := []g.Node{}
		childComments := hasChildren.AllChildren()
		for i := len(childComments) - 1; i >= 0; i-- {
			child := childComments[i]
			if child, ok := child.(Comment); ok {
				children = append(children, CommentWithChildren(child))
			}
		}
		return g.Group(append([]g.Node{comment}, Div(Class("ml-1"), g.Group(children))))
	}

	return comment
}

func CommentBlock(c Comment) g.Node {
	return Div(
		Class("flex flex-col space-y-2 border-l-2 mt-1 pl-2 border-orange-700"),
		P(Class("font-mono text-sm"), g.Textf("%s at ", c.CommentAuthor()), TimeLabel(c.WrittenAt())),
		P(Class("prose-lg prose-stone"), g.Text(c.CommentContent())),
		CommentLink(c.CommentableID()),
	)
}
