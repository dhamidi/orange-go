package pages

import (
	g "github.com/maragudk/gomponents"

	. "github.com/maragudk/gomponents/html"
)

func SubmitPage(path string, form *FormState, context *PageData) g.Node {
	return Page("The Orange Website | Submit", path, SubmitForm(form), context)
}

func SubmitForm(form *FormState) g.Node {
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		Div(
			Class("sm:mx-auto sm:w-full sm:max-w-sm"),
			H2(Class("mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"),
				g.Textf("Submit a link"))),
		Div(
			Class("mt-10 sm:mx-auto sm:w-full sm:max-w-sm"),
			Form(Class("space-y-6"), Action("/submit"), Method("POST"),
				InputWithLabel("url", "URL", "text", form, Required()),
				InputWithLabel("title", "Title", "text", form, Required()),
				SubmitButton("Submit"),
			),
		),
	)
}
