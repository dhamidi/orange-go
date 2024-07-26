package pages

import (
	g "github.com/maragudk/gomponents"

	. "github.com/maragudk/gomponents/html"
)

func LoginPage(path string, context *PageData) g.Node {
	return Page("The Orange Website | Login", path, LoginForm(), &PageData{})
}

func LoginForm() g.Node {
	form := NewFormState()
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		Div(
			Class("sm:mx-auto sm:w-full sm:max-w-sm"),
			H2(Class("mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"),
				g.Textf("Welcome"))),
		Div(
			Class("mt-10 sm:mx-auto sm:w-full sm:max-w-sm"),
			Form(Class("space-y-6"), Action("/login"), Method("POST"),
				InputWithLabel("username", "Username", "text", form, Required()),
				InputWithLabel("password", "Password", "password", form, Required(), g.Attr("autocomplete", "current-password")),
				SubmitButton("Log in"),
			),
		),
	)
}
