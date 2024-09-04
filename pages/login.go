package pages

import (
	"net/url"

	g "github.com/maragudk/gomponents"

	. "github.com/maragudk/gomponents/html"
)

func LoginPage(path string, context *PageData) g.Node {
	return Page("The Orange Website | Login", path, LoginForm(context.FormState, context.BackTo), context)
}

func LoginForm(form *FormState, backTo *url.URL) g.Node {
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		Div(
			Class("sm:mx-auto sm:w-full sm:max-w-sm"),
			H2(Class("mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"),
				g.Textf("Welcome"))),
		Div(
			Class("mt-10 mb-2 sm:mx-auto sm:w-full sm:max-w-sm"),
			Form(Class("space-y-6 mb-1"), Action("/login"), Method("POST"),
				g.Iff(backTo != nil, func() g.Node {
					return Input(Type("hidden"), Name("back_to"), Value(backTo.String()))
				}),
				InputWithLabel("username", "Username", "text", form, Required()),
				InputWithLabel("password", "Password", "password", form, Required(), g.Attr("autocomplete", "current-password")),
				SubmitButton("Log in"),
			),
			ButtonLink("✨ Use magic ✨", "/magic"),
		),
		Div(
			Class("mt-1 border-t border-gray-700 sm:mx-auto sm:w-full sm:max-w-sm pt-1"),
			ButtonLink("⚙️  Reset your password ⚙️ ", "/reset-password"),
		),
	)
}
