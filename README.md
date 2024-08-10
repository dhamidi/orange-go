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

Commands can be added to the command log through multiple ways:

1. modifying the command log database directly,
2. using the `orange` command,
3. performing actions on the web interface.

Before serving a request, the web application checks for new commands
and applies them.

Thus changes made using the `orange` command take effect immediately
on the next web request.

### Skipping commands

Using the `orange` command, entries in the command log can be masked
to instruct the application to skip over them.to the state without
losing any data long term.

This change can be reverted by issuing another `orange` command.

In case the application has entered an undesirable state, skipping
commands allows quickly making changes to the state without losing
any data long term.

```shell
# skip command log entries 3 and 5
./orange skip-commands 3 5

# revert the operation
./orange unskip-commands 3 5
```

## Module: Auth

The authentication module is responsible for everything related to user accounts.

Users are identified by unique usernames.

The username is assigned when a user first signs up and cannot be _changed later_.

Which usernames are allowed is controlled by a username policy,
limiting new usernames by minimum and maximum length, as well as
disallowing certain usernames.

The default username policy enforces no length minimum, and a maximum of 32 characters.

**TODO**: allow changing password on behalf of the users.

**TODO**: allow users to reset their password.

## Module: Content

The heart of the application, the content module is responsible for
submissions and comments.

Submissions consist of a URL and a user-submitted title.

A background goroutine fetches OpenGraph data for the URL and stores new submissions.

Submissions can be commented on, and upvoted.

**TODO**: make upvotes affect the sorting of submissions
