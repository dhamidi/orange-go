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

### Admin users

A set of users can be designated as administrators.

Administrators have access to a set of endpoints under `/admin`,
which allows the to inspect the state of the system and perform
certain actions.

### Magic links

Users that have a verified email address or have access to an email
address belonging to a magic domain, can log in using a magic link:
a link submitted via email that allows them to log in directly.

To configure the list of magic domains, run the `set-magic-domains` command:

```shell
# allow everyone with a @bolt.eu email address to log in using magic links
./orange set-magic-domains bolt.eu
```

When a user logs in for the first time with a magic link, a new user account
is created: the part of the email address before the `@` symbol is used
as the user's username.

### Verified email addresses

There are two ways to verify an email address:

1. an administrator manually links an already verified address:
   `./orange link-email <username> <email>`
2. a user verifies their email address by clicking a link sent to the email address

**TODO**: allow changing password on behalf of the users.

**TODO**: allow users to reset their password.

## Module: Content

The heart of the application, the content module is responsible for
submissions and comments.

Submissions consist of a URL and a user-submitted title.

A background goroutine fetches OpenGraph data for the URL and stores new submissions.

Submissions can be commented on, and upvoted.

Upvoted submissions are shown in the order of their score.

Scoring is based on number of upvotes, decaying over time.

Submissions are re-scored every time a new link is submitted or an
existing one is upvoted.
