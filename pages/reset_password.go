package pages

import (
	g "github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	. "github.com/maragudk/gomponents/html"
)

func ResetPasswordPage(path string, context *PageData) g.Node {
	return Page("The Orange Website | Reset Password", path, ResetPasswordForm(context.FormState), context)
}

func ResetPasswordForm(form *FormState) g.Node {
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		Div(
			H2(
				Class("mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"),
				g.Textf("Request a new password"),
			),
		),
		Div(
			Class("mt-10 sm:mx-auto sm:w-full sm:max-w-sm"),
			Form(
				hx.Post("/reset-password"),
				hx.Swap("outerHTML"),
				Class("space-y-6"),
				Action("/reset-password"),
				Method("POST"),
				InputWithLabel("email", "Email", "email", form, Required()),
				SubmitButton("Send password reset link"),
			),
		),
	)
}

func ResetPasswordSentPage(path string, context *PageData) g.Node {
	return Page(
		"The Orange Website | Password Request sent",
		path,
		ResetPasswordSent(),
		context,
	)
}
func ResetPasswordSent() g.Node {
	return Div(
		Class("flex flex-row justify-center items-center py-12"),
		P(Class("text-xl font-semibold"), g.Textf("Password reset request sent ⚙️ ")),
	)
}
