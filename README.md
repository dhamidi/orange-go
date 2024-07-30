# Description

A HackerNews clone written to try out vertical scaling.

Ideas found here:

- a simple stack: Go, sqlite, HTMX, TailwindCSS
- log-based architecture: derive all state from a message log
- modules/services living _within the same process_, i.e. microservices without
  the network boundaries, or rather bounded contexts from Domain Driven Design.

## Features

- User signup and login,
- Posting submissions,
- Upvoting submissions,
- Commenting on submissions and on submission comments (no limit to nesting)
- Fetching OpenGraph data for submitted URLs

## Operation

When the application starts, the log is replayed to derive the state.

A failure to replay a message results in the application not starting.

Messages are persisted in a sqlite3 database in the file `commands.db`

Only inputs accepted by the system are persisted, there is no distinction
between commands and events.
