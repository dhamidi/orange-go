package pages

import (
	"fmt"
	"reflect"

	g "github.com/maragudk/gomponents"

	. "github.com/maragudk/gomponents/html"
)

type EventLogEntry struct {
	ID      string
	Kind    string
	Payload interface{}
}

func EventLogPage(path string, data []*EventLogEntry, context *PageData) g.Node {
	return Page("The Orange Website | Event Log", path, EventList(data), context)
}

func EventList(logEntries []*EventLogEntry) g.Node {
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		g.Group(g.Map(logEntries, renderEvent)),
	)
}

func renderEvent(logEntry *EventLogEntry) g.Node {
	return Div(Class("font-mono flex flex-row"),
		Span(Class("min-w-16 mr-2 text-right inline-block"), g.Text(logEntry.ID)),
		Div(
			Details(
				Summary(g.Text(logEntry.Kind)),
				renderEventFields(logEntry.Payload),
			),
		),
	)
}

func renderEventFields(payload any) g.Node {
	fields := []g.Node{}
	addField := func(f g.Node) {
		fields = append(fields, f)
	}

	p := reflect.ValueOf(payload).Elem()
	for i := 0; i < p.NumField(); i++ {
		addField(Div(
			Span(Class("mr-2 w-48 text-clip text-right inline-block"), g.Text(p.Type().Field(i).Name)),
			Span(Class("text-left inline-block"), g.Text(formatFieldValue(p.Field(i)))),
		))
	}

	return g.Group(fields)
}
func formatFieldValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Struct:
		if t, ok := v.Interface().(fmt.Stringer); ok {
			return fmt.Sprintf("%q", t)
		}
		if v.CanAddr() {
			if t, ok := v.Addr().Interface().(fmt.Stringer); ok {
				return fmt.Sprintf("%q", t)
			}
		}
		return fmt.Sprintf("%#v", v.Interface())
	case reflect.Pointer:
		if v.IsZero() {
			return "<nil>"
		} else {
			return fmt.Sprintf("%q", v.Elem())
		}
	default:
		return fmt.Sprintf("%q", v.Interface())
	}
}
