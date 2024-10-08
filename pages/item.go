package pages

import (
	"fmt"
	"strings"

	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"

	. "github.com/maragudk/gomponents/html"
)

func ItemPage(path string, submission *Submission, context *PageData) g.Node {
	detail := WithoutCommentForm
	if context.CurrentUser != nil {
		detail = WithCommentForm
	}
	return Page("The Orange Website", path, SubmissionDetail(submission, detail, context.IsAdmin), context)
}

type SubmissionDetailElement string

const (
	WithoutCommentForm SubmissionDetailElement = "without_comment_form"
	WithCommentForm    SubmissionDetailElement = "with_comment_form"
)

func SubmissionDetail(s *Submission, with SubmissionDetailElement, isAdmin bool) g.Node {
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
				g.If(with == WithCommentForm, CommentForm(s.ItemID, NewFormState())),
				g.If(with == WithoutCommentForm, ButtonLink("Log in to comment", href("/login", q{"back_to": "/item?id=" + s.ItemID}))),
			),
		),
		g.Group(g.Map(s.Comments, func(c Comment) g.Node {
			return CommentWithChildren(c, isAdmin)
		})),
	)
}

func CommentParent(itemID string) g.Node {
	return A(
		Href("/item?id="+itemID),
		Span(Class("text-xs mx-1 font-mono"), g.Text("[parent]")),
	)
}

func CommentLink(itemID string, formTarget string) g.Node {
	return A(
		hx.Get("/comment?itemID="+itemID),
		hx.Swap("outerHTML"),
		hx.Target(formTarget),
		Href("/item?id="+itemID),
		Span(Class("text-xs mx-1 font-bold font-mono"), g.Text("[reply]")),
	)
}

func CommentForm(itemID string, state *FormState) g.Node {
	return Form(
		g.Attr("x-data", `{ n: 0 }`),
		hx.Post("/comment"),
		hx.Swap("this"),
		Action("/comment"), Method("POST"),
		Div(
			Class("flex flex-row space-around"),
			Input(Type("hidden"), Name("itemID"), Value(itemID)),
			InlineText("text", state, g.Attr("x-on:keyup", `n = $event.target.value.length`)),
			InlineSubmitButton("Reply"),
		),
		P(Class("text-xs text-gray-400 my-1"), g.Attr("x-text", `(3000 - n) + " characters left"`)),
	)
}

func CommentWithChildren(c Comment, isAdmin bool) g.Node {
	comment := CommentBlock(c, isAdmin)
	if hasChildren, ok := c.(WithChildren); ok {
		children := []g.Node{}
		childComments := hasChildren.AllChildren()
		for i := len(childComments) - 1; i >= 0; i-- {
			child := childComments[i]
			if child, ok := child.(Comment); ok {
				children = append(children, CommentWithChildren(child, isAdmin))
			}
		}
		return g.Group(append([]g.Node{comment}, Div(Class("ml-1"), g.Group(children))))
	}

	return comment
}

func CommentBlock(c Comment, isAdmin bool) g.Node {
	commentFormTarget := strings.Replace(fmt.Sprintf("c-%s", c.CommentableID()), "/", "-", -1)
	return Div(
		Class("flex flex-col text-xs font-mono border-l-2 mt-1 pl-2 border-orange-700"),
		Div(Class("text-xs"),
			A(
				Class("cursor-pointer"),
				Href(href("/item", q{"id": c.CommentableID()})),
				g.Textf("%s at ", c.CommentAuthor()),
				TimeLabel(c.WrittenAt())),
			CommentParent(c.CommentParentID()),
			CommentLink(c.CommentableID(), "#"+commentFormTarget),
			CommentAdminActions(isAdmin, c),
		),
		Div(Class("prose text-xs my-1 prose-stone"), g.Raw(c.CommentContent())),
		Div(ID(commentFormTarget)),
	)
}

func CommentAdminActions(isAdmin bool, c Comment) g.Node {
	return g.If(
		isAdmin,
		g.Group([]g.Node{
			g.Text(" | "),
			g.If(c.IsHidden(), UnhideCommentButton(c.CommentID())),
			g.If(!c.IsHidden(), HideCommentButton(c.CommentID())),
		}),
	)
}
