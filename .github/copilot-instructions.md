# GitHub Copilot Instructions

This file provides rules for GitHub Copilot when reviewing pull requests in `goravel/framework`.

## Go Code Style

- Use `any` instead of `interface{}`.
- Do not shadow the built-in `error` package or common variable names (`err`, `ctx`, etc.).
- Import order: standard library → third-party → internal (`github.com/goravel/framework/...`), separated by blank lines.

## Error Handling

- All sentinel errors must be declared in `errors/list.go` using `errors.New(...)`. Do not create ad-hoc `fmt.Errorf` errors for domain errors.
- Tag errors with a module via `.SetModule(errors.ModuleXxx)` when returning from a service provider or internal package.
- Use `%s`/`%v` format verbs in error messages and supply args via `.Args(...)` — never interpolate directly into the `New` string.
- Do not swallow errors silently; log or propagate them.

## Testing

See [.agents/prompts/tests.md](../.agents/prompts/tests.md) for the full testing guidelines.
