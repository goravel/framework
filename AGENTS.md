## Project Overview

Goravel (`github.com/goravel/framework`) is the framework core package, not an application.
Application scaffold: [goravel/goravel](https://github.com/goravel/goravel).

## Common Commands

```bash
go test                                # unit tests
cd tests && go test ./...              # integration tests
go tool mockery                        # regenerate mocks
golangci-lint run                      # lint
```

## Core Architecture

- Main pattern: `contracts/` -> module implementation -> `facades/facades.go`.
- Each module uses `service_provider.go` with:
	- `Register(app)` for bindings
	- `Boot(app)` for post-registration setup
	- `Relationship()` for dependency order (`ServiceProviderWithRelations`)
- IoC container: `foundation/container.go` (`Bind`, `Singleton`, `BindWith`, `Instance`, `Make`).
- Binding registry: `contracts/binding/binding.go`.

## Key Directories

- `contracts/`: public interfaces.
- `foundation/`: app bootstrap, container, provider repository.
- `facades/`: static-style service access.
- `mocks/`: generated `testify` mocks.
- `support/`: shared utilities.
- `testing/`: test helpers and framework.
- `tests/`: integration test module.
- `packages/`: package development support.

## External Drivers

- Route: `goravel/gin`, `goravel/fiber`
- Database/ORM: `goravel/mysql`, `postgres`, `sqlite`, `sqlserver`
- Cache/Session/Queue: `goravel/redis`
- Storage: `goravel/s3`, `oss`, `cos`, `minio`

## Planning Rules

- When asked to plan or investigate, produce a plan only — do not implement it until the user explicitly asks you to proceed.

## Code Rules

- Use `any` instead of `interface{}`.
- Never edit `mocks/` directly; run `go tool mockery` to regenerate.
- Follow standard Go formatting/naming; add comments where logic isn't self-evident. Go version is in go.mod.
- Never create errors inline in logic (`fmt.Errorf`, `errors.New`, etc.). Declare all errors as named variables in `errors/list.go` using the framework's `New(...)` constructor. This centralises error messages to support future i18n.

## Tests

When writing/running tests, use the rules in `.agents/prompts/tests.md` for guidance.
