# GitHub Copilot Instructions

This file provides rules for GitHub Copilot when reviewing pull requests in `goravel/framework`.

## Go Code Style

- Use `any` instead of `interface{}`.
- Do not shadow the built-in `error` type, the standard-library `errors` package, or common variable names (`err`, `ctx`, etc.).
- Import order: standard library → third-party → internal (`github.com/goravel/framework/...`), separated by blank lines.

## Error Handling

- Never create errors inline in logic. All errors — including one-off messages — must be declared as named variables in `errors/list.go` using the framework error constructor: `github.com/goravel/framework/errors.New(...)` (or unqualified `New(...)` when inside `package errors`). This applies to every package; do not use `fmt.Errorf`, `errors.New`, or any stdlib error constructor at the call site. Errors are centralised in `errors/list.go` to support future i18n.
- Tag errors with a module via `.SetModule(errors.ModuleXxx)` when returning from a service provider or internal package.
- Use format verbs in error messages and supply dynamic parts via `.Args(...)` — never interpolate directly into the `New` string.
- Do not swallow errors silently; log or propagate them.

## Testing

See [.agents/prompts/tests.md](../.agents/prompts/tests.md) for the full testing guidelines.
