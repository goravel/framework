---
name: make-plan
description: >
  Planning workflow for Goravel tasks. Use this skill whenever the user asks
  for a plan, investigation, or implementation outline and the response should
  show both the concrete code change and the real user-facing code that the
  change enables.
---

# Make Plan

Produce plans that are specific enough for the user to evaluate before any code
is written.

## Core rule

Every plan must show two different views when code is involved:

- **Actual code**: the concrete framework or internal code you expect to change.
- **Real user code**: the end-user application code that would use or observe the
  change.

Do not stop at abstract steps like "update middleware" or "add validation".
Show the shape of the real code.

## Planning rules

1. Follow the repository rule: when asked to plan or investigate, produce a plan
   only. Do not implement until the user explicitly asks.
2. Keep the plan grounded in the current codebase. Read the relevant files before
   proposing code.
3. For every non-trivial step, explain which file or component will likely change.
4. If the work affects user-facing behavior, include a `Real user code` example.
5. If the work is entirely internal, say that explicitly and omit the user code
   example only when no realistic user-facing snippet exists.
6. Never use pseudo-code placeholders inside code fences. Use realistic Goravel
   code patterns.

## Output format

When a plan includes code changes, use this structure:

````markdown
## Plan
1. <step>
2. <step>
3. <step>

## Actual code
```go
// concrete framework code shape expected to change
```

## Real user code
```go
// concrete application code a Goravel user would write
```
````

If multiple steps affect different areas, include multiple paired code examples.
Keep each example short and representative.

## What counts as actual code

Use `Actual code` to show the likely implementation shape in this repository,
such as:

- contracts added under `contracts/`
- framework wiring in `foundation/` or `service_provider.go`
- facade or module behavior in package code
- tests that prove the framework behavior

Example:

```go
type Middleware interface {
	Handle(ctx http.Context) error
	Name() string
}
```

## What counts as real user code

Use `Real user code` to show code a Goravel application author would actually
write in a real app, not internal framework implementation.

Prefer patterns from `goravel/goravel` and `goravel/example`, such as:

- route registration
- middleware usage
- controller code
- request validation
- ORM queries
- auth flows

Example:

```go
facades.Route().Middleware(middleware.Jwt()).Get("users", userController.Index)
```

## Good vs bad plans

Bad:

````markdown
## Plan
1. Add middleware support.
2. Update tests.
````

Good:

````markdown
## Plan
1. Extend the route pipeline builder so route groups can prepend default
   middleware before per-route middleware is appended.
2. Update the HTTP contract and route tests to verify middleware ordering.
3. Validate the user-facing API still reads naturally from an application route
   file.

## Actual code
```go
func (r *Router) UseDefaultMiddleware(middlewares ...http.Middleware) {
	r.defaultMiddlewares = append(r.defaultMiddlewares, middlewares...)
}
```

## Real user code
```go
facades.Route().Prefix("api").Group(func(route route.Router) {
	route.Middleware(middleware.Jwt()).Get("profile", userController.Show)
})
```
````

## Workflow

1. Read the relevant framework files.
2. Identify the behavior change the user cares about.
3. Write a short step-by-step plan.
4. Add `Actual code` showing the concrete framework change shape.
5. Add `Real user code` showing how an app developer would use it.
6. Call out unknowns or risks if the code path is ambiguous.

## Guardrails

- Never present only implementation detail when the feature affects application
  authors.
- Never present only user-facing code without showing what framework area will
  change.
- Never use fake APIs that do not match Goravel naming and style.
- Prefer concise, concrete examples over long explanations.
