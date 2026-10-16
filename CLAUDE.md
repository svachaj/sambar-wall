# sambar-wall

Go web app: Echo + templ + HTMX + Alpine + Tailwind v3, on the legacy MSSQL database. The UI is Czech only.

## Commands

- `make templ`: regenerates `*_templ.go`. Run it after editing any `.templ` file and commit the generated files.
- `make test`: Go unit tests.
- `make e2e`: Playwright end-to-end suite (see README, "End-to-end tests").

## End-to-end tests (e2e/)

- **Every user-facing change adds or updates a spec in `e2e/tests/`.** The pre-push hook (`.githooks/pre-push`, enabled by `make install-hooks`) runs `make e2e` on every push.
  - Never bypass a failing run with `SKIP_E2E=1` or `--no-verify` unless the user asks.
  - When pushing from Claude Code, run `git push` outside the sandbox, because the hook needs the DB.
- **Deployed smoke tests:** `make e2e-deployed` (`e2e/deployed/`) checks https://registrace-test.stenakladno.cz, which uses the same DB.
  - The deployed app sends real emails. Never add a deployed test that requests a login or agreement code, submits an application or toggles "paid".
  - Log in with `loginWithoutEmail()`, and intercept downloads.
- **The suite runs against the shared test DB from `.env`.**
  - Only one run at a time: `e2e/lib/lock.ts` holds the `sambar-wall-e2e` application lock from setup until cleanup is done. Keep any new setup or teardown inside `global-setup.ts`, under that lock.
  - Create data only through `e2e/lib/seed.ts`, or with `@e2e-sambar.test` emails, so `purge()` can remove it.
  - Never run `POST /prihlasky-opakovat` (it bulk-creates real applications).
- **Selectors:**
  - Prefer `getByRole` and `getByLabel` with the Czech labels.
  - Use `tick()` for checkboxes, because htmx re-renders them on click.
  - Call `htmxSettled(page)` before interacting with content htmx just swapped in, because its triggers are attached only after settling.
  - Use `toast(page, text)` for toasts.
  - Use `row(page, text)` for table rows.
- **Known bugs** are tests marked `test.fail(true, '<reason>')`. Remove the marker once the bug is fixed.
- **Sandbox:** the DB is a named SQL Server instance (`host\SQLEXPRESS`), resolved over UDP via SQL Browser. Commands that touch the DB (`make e2e`, `go run .`) need network access outside the sandbox.
