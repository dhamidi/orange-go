package pages

import (
	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
)

func RequestMagicPage(path string, context *PageData) g.Node {
	return Page("The Orange Website | Login", path, RequestMagicForm(context.FormState), context)
}

func RequestMagicForm(form *FormState) g.Node {
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		Div(
			H2(
				Class("mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"),
				g.Textf("Use your email to get a magic link"),
			),
		),
		Div(
			Class("mt-10 sm:mx-auto sm:w-full sm:max-w-sm"),
			Form(
				hx.Post("/magic"),
				hx.Swap("outerHTML"),
				Class("space-y-6"),
				Action("/magic"),
				Method("POST"),
				InputWithLabel("email", "Email", "email", form, Required()),
				SubmitButton("Send magic link"),
			),
		),
	)
}

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

func MagicSentPage(path string) g.Node {
	return Div(
		Class("flex flex-row justify-center items-center py-12"),
		P(Class("text-xl font-semibold"), g.Textf("Magic has been sent 🧙🦉")),
	)
}
