# Web application: Sambar Lezecká Stěna - Kroužky a registrace

https://registrace.stenakladno.cz/

This application is designed for registration for climbing wall courses and camps in Kladno. It serves as a pilot project from which I would like to develop an open-source registration platform for courses and similar events. For now, it is available only in Czech localization.

## Important note

This application uses MSSQL as its database because it is a replacement for an existing interface to an older version of the application and is therefore tightly bound to MSSQL. Based on this application, a general template will be created, which should include an ORM to make everything more flexible.

The application is currently strictly localized to the Czech language, which is another aspect to be improved in the general template.

### The result should be a preconfigured template for developing modern and high-performance web applications based on Go and HTMX, with various authentication options.

## Auth

This application uses authentication via a so-called magic email link. Nowadays, authentication with just a username and password is not very secure. This type of authentication provides a higher level of security while also being extremely user-friendly, which is a crucial advantage.

Email clients today are inherently secure enough, so shifting the responsibility to them makes sense. While this method is not strictly two-factor authentication in the traditional sense, it certainly offers a much stronger security model than a simple username and password.

I highly recommend this authentication approach, and this repository can serve as a full reference for implementation and inspiration.

[Security handler](modules/security/security-handlers.go)

## Prerequisites

### [Go](https://go.dev/) of course 🚀

See go mod file for the version of Go used in this project.

### Makefile command to install all the other tools needed for the project

`make install-tools` (for Linux x64)

### [TEMPL](https://templ.guide/) - Template Engine

To have the best templ experience, you should install VSCode extension `templ-vscode`.

Than follow this link how to setup IDE [IDE Support](https://templ.guide/commands-and-tools/ide-support)

### [Tailwindcss](https://tailwindcss.com/) - CSS Framework

With no need to install npm, you can use the standalone CLI to compile your Tailwind CSS files.
`https://tailwindcss.com/blog/standalone-cli`

To have the best tailwindcss experience, you should install VSCode extension `Tailwind CSS IntelliSense` and let the extension to know where the tailwindcss is located, because we are using .templ files instead of .html files.

Extension settings:
`"tailwindCSS.includeLanguages": {
  "templ": "html"
}`

### [Air](https://github.com/air-verse/air) - Live reload for Go apps

## How to develop

### 1. Clone the repository

### 2. Set up prerequisites

### 3. Create .env file in the root of the project to set up the environment variables for the configuration

`APP_PORT=5500`

### 3. Run with Air

`air`

## End-to-end tests

The `e2e/` folder holds a [Playwright](https://playwright.dev/) suite. It starts the real Go app and clicks through every flow: login (code and magic link), the wall agreement, course applications, the admin application list, attendance, wall visitors and role-based access.

### Setup (once)

1. Install Node.js 20 or newer.
2. Run `make e2e-install`.
3. Run `cp e2e/.env.example e2e/.env` and set `E2E_ALLOW_DB` to the name of the test database.

`make e2e-install` also runs `make install-hooks`, which enables the [pre-push hook](.githooks/pre-push). Every `git push` then runs the suite first and stops if a test fails. Pushes that change only Markdown files skip the suite. To skip it once, use `SKIP_E2E=1 git push` (or `git push --no-verify`).

### Run

| Command | What it does |
| --- | --- |
| `make e2e` | Runs the whole suite (about 40 s) |
| `make e2e-ui` | Runs it in Playwright's interactive UI mode |
| `make e2e-report` | Opens the HTML report of the last run, with traces of failed tests |

Playwright builds the app into `tmp/e2e-server` and starts it on port 5599, using the database from the root `.env`. It forces `APP_ENV=localhost`, so no emails are sent, and `PAYMENTS_CRON=0`, so no payment emails are read. Tests read login and agreement codes straight from `t_system_registration_code`.

### Shared test database

The suite writes to the database configured in `.env`, so it has some safeguards:

- **Only an explicitly allowed database.** It refuses to start unless `DB_NAME` equals `E2E_ALLOW_DB`.
- **Every created account or visitor uses the `@e2e-sambar.test` domain.**
- **Clearly marked test courses.** The two test courses are cloned from an existing course and marked by `TimeTo = 23:59` and `ValidTo = 2099-12-31`.
- **Automatic cleanup.** Before and after each run, everything tied to those accounts and courses is deleted. The run fails if anything is left behind.
- **Not tested:** the bulk action "Poslat hromadně přihlášky znovu" (`POST /prihlasky-opakovat`). It would re-create applications for every "will continue" record in the database.

Set `E2E_KEEP_DATA=1` to keep the test data after a run for debugging; the next run removes it. Set `E2E_SERVER_LOGS=1` to print the app's log.

Tests marked `test.fail(...)` document known bugs. When a bug gets fixed, its test starts "unexpectedly passing": remove the `test.fail` line at that point.
