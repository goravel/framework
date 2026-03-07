## Project Overview

Goravel (`github.com/goravel/framework`) is the framework core package, not an application.
Application scaffold: [goravel/goravel](https://github.com/goravel/goravel).

## Common Commands

```bash
go test ./...                          # all unit tests
go test ./cache/...                    # package tests
go test ./cache/... -run TestMemory    # single test
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

## Build Flow

`foundation.Application.Build()` handles configuration, provider registration/boot, middleware/events/commands setup, and command registration.

## AI Agent Go Style Rule

- When writing Go code, AI agents should use `any` instead of `interface{}`.
- Avoid adding `mock.Anything` when writing test cases.
- Use the testify `EXPECT` method when writing test cases.

### Examples

1. Use `any` instead of `interface{}`

```go
// Avoid
func Handle(data interface{}) error {
	return nil
}

// Prefer
func Handle(data any) error {
	return nil
}
```

2. Avoid `mock.Anything` in tests

```go
// Avoid
repo.EXPECT().Find(mock.Anything).Return(user, nil)

// Prefer
repo.EXPECT().Find("user-1").Return(user, nil)
```

3. Use `EXPECT` method for mock expectations

```go
// Avoid
repo.On("Find", "user-1").Return(user, nil)

// Prefer
repo.EXPECT().Find("user-1").Return(user, nil)
```
