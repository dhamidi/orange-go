package pages

import (
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
	renderEvent := func(logEntry *EventLogEntry) g.Node {
		return Span(Class("font-mono block"), g.Text(logEntry.Kind))
	}
	return Div(
		Class("flex min-h-full flex-col justify-center px-6 py-12 lg:px-8"),
		g.Group(g.Map(logEntries, renderEvent)),
	)
}
