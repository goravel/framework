
# Tests

- Prefer `go test <pkg>` or `-run <TestName>` over `go test ./...` (slow).
- Use table-driven tests covering happy path, failure, and edge cases.
- Skip trivial getters/setters unless they contain non-trivial logic.
- Use `testify/assert` with `assert.*(t, *)` or `require.*(t, *)` directly, not `assert.New(t)`.
- Use `testify/suite` for related test groups, and use `s.*(*, *)` when asserting.
- Use the testify `EXPECT` method for mocks; avoid `mock.Anything`.
- Use `assert.AnError` if needed, and `assert.Equal` for error assertions.
- Assert full maps/structs/slices/arrays, not individual fields.
- Prefer direct value assertions over `mock.MatchedBy`; use it only for dynamically-generated args.
- Every mock must use `.Once()` or `.Times()`, only use `.Maybe()` when necessary; avoid no-op expectations.
- Name tests `Test<FunctionName>_[Optional]`; use table style or sub-tests for multiple cases.
- Don't use `assert.*` with `if` statements; use `assert.*` directly for clarity and better failure messages.
- Use `t.Run()` for sub-tests when testing multiple cases for the same function, and use table-driven tests for multiple cases with similar setup/assertions. Avoid writing separate test functions for each case when they share common logic.
- The basic table-driven test pattern is:

```go
import (
	[system packages]

	[third-party packages]

	[internal packages]
)

func TestFunction(t *testing.T) {
	// The name should start with `mock` to indicate it's a mocked function.
	var (
		ctx 	context.Context
		mockFunc *mocks.MockedInterface
	)

	beforeEach := func() {
		mockFunc = mocks.NewMockedInterface(t)
	}

	tests := []struct {
		name        string
		input       any
		setup       func()
		expect      any
		expectError error
	}{
		{
			name:  "should do something",
			input: someInput,
			setup: func() {
				mockFunc.EXPECT().SomeMethod(someArgs).Return(someResult, nil).Once()
			},
			expect:      someResult,
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()

			result, err := FunctionUnderTest(tt.input)
			assert.Equal(t, tt.expect, result)
			assert.Equal(t, tt.expectError, err)
		})
	}
}
```
