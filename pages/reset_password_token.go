package pages

import (
	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
)

func ResetPasswordTokenPage(path string, context *PageData) g.Node {
	return Page("The Orange Website | Reset Password", path, ResetPasswordTokenForm(context.FormState), context)
}

func ResetPasswordTokenForm(form *FormState) g.Node {
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		Div(
			H2(
				Class("mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"),
				g.Textf("Set your new password"),
			),
		),
		Div(
			Class("mt-10 sm:mx-auto sm:w-full sm:max-w-sm"),
			Form(
				hx.Post("/reset-password/"+form.Values["token"]),
				hx.Swap("outerHTML"),
				Class("space-y-6"),
				Action("/reset-password/"+form.Values["token"]),
				Method("POST"),
				Input(Type("hidden"), Name("token"), Value(form.Values["token"])),
				InputWithLabel("password", "New password", "password", form, Required()),
				SubmitButton("Reset password"),
			),
		),
	)
}

func ResetPasswordSuccessPage(path string, context *PageData) g.Node {
	return Page(
		"The Orange Website | Password Reset",
		path,
		ResetPasswordSuccess(),
		context,
	)
}
func ResetPasswordSuccess() g.Node {
	return Div(
		Class("flex flex-col justify-center items-center py-12"),
		P(Class("block text-xl font-semibold"), g.Textf("Your password has been reset ⚙️ ")),
		P(
			Class("block mt-4"),
			ButtonLink("Log in now", "/login"),
		),
	)
}
