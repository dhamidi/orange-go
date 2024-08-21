package pages

import (
	"strings"

	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"

	. "github.com/maragudk/gomponents/html"
)

type FormState struct {
	Errors map[string]string
	Values map[string]string
}

func NewFormState() *FormState {
	return &FormState{
		Errors: map[string]string{},
		Values: map[string]string{},
	}
}

func (s *FormState) AddError(field, message string) *FormState {
	s.Errors[field] = message
	return s
}

func (s *FormState) HasErrors() bool {
	return len(s.Errors) > 0
}

func (s *FormState) SetValue(field, value string) *FormState {
	s.Values[field] = value
	return s
}

func (s *FormState) ErrorFor(field string) string {
	if s == nil {
		return ""
	}
	return s.Errors[field]
}

func (s *FormState) HasErrorFor(field string) bool {
	if s == nil {
		return false
	}
	_, ok := s.Errors[field]
	return ok
}

func (s *FormState) ValueFor(field string) g.Node {
	if s == nil {
		return Value("")
	}
	return Value(s.Values[field])
}

func InlineSubmitButton(label string) g.Node {
	return Span(Class("mx-2"),
		Button(Type("submit"), Class("flex w-full justify-center rounded-md bg-orange-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-orange-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange-600"),
			g.Textf(label)),
	)

}
func SubmitButton(label string) g.Node {
	return Div(
		Div(Class("mt-2"),
			Button(Type("submit"), Class("flex w-full justify-center rounded-md bg-orange-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-orange-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange-600"),
				g.Textf(label)),
		),
	)
}

func ButtonLink(label, path string) g.Node {
	return A(
		Href(path),
		Class("flex w-full justify-center rounded-md bg-orange-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-orange-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange-600"),
		g.Textf(label),
	)
}

func InputWithLabel(name, label, kind string, state *FormState, attrs ...g.Node) g.Node {
	classes := "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-orange-600 sm:text-sm sm:leading-6"
	return Div(
		Label(For(name), Class("block text-sm font-medium leading-6 text-gray-900"),
			g.Textf(label)),
		Div(Class("mt-2"),
			Input(append([]g.Node{
				ID(name),
				Name(name),
				Type(kind),
				state.ValueFor(name),
				c.Classes{
					classes: !state.HasErrorFor(name),
					strings.ReplaceAll(classes, "ring-gray-300", "ring-red-500"): state.HasErrorFor(name),
				},
			},
				attrs...,
			)...,
			),
			g.If(state.HasErrorFor(name), P(Class("text-sm text-red-400"), g.Text(state.ErrorFor(name)))),
		),
	)
}

func InlineText(name string, state *FormState, attrs ...g.Node) g.Node {
	classes := "block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-orange-600 sm:text-sm sm:leading-6"
	return Div(Class("inline-block flex-auto"),
		Div(Class("mx-2"),
			Textarea(append([]g.Node{
				ID(name),
				Name(name),
				Type("text"),
				state.ValueFor(name),
				c.Classes{
					classes: !state.HasErrorFor(name),
					strings.ReplaceAll(classes, "ring-gray-300", "ring-red-500"): state.HasErrorFor(name),
				},
			},
				attrs...,
			)...,
			),
			g.If(state.HasErrorFor(name), P(Class("text-sm text-red-400"), g.Text(state.ErrorFor(name)))),
		),
	)
}
