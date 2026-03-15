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

## AI Agent Code Rule

- Should use `any` instead of `interface{}`.
- Avoid adding `mock.Anything` when writing test cases.
- Use the testify `EXPECT` method when writing test cases.
- Don't modify the files in the `mocks` directory, run the `go tool mockery` command to regenerate mocks instead if needed.
- Don't run `go test ./...` if unnecessary, the command is a bit slow, run `go test` with the specific package or test function instead.