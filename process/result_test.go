package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultMethods_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		res            *Result
		expectSuccess  bool
		expectFailed   bool
		expectExitCode int
		expectOutput   string
		expectErrOut   string
		expectCommand  string
		seeOut         string
		seeErr         string
		seeOutWant     bool
		seeErrWant     bool
	}{
		{
			name:           "nil receiver",
			res:            nil,
			expectSuccess:  false,
			expectFailed:   true,
			expectExitCode: -1,
			expectOutput:   "",
			expectErrOut:   "",
			expectCommand:  "",
			seeOut:         "anything",
			seeErr:         "anything",
			seeOutWant:     false,
			seeErrWant:     false,
		},
		{
			name:           "successful result",
			res:            NewResult(0, "echo hi", "hi", ""),
			expectSuccess:  true,
			expectFailed:   false,
			expectExitCode: 0,
			expectOutput:   "hi",
			expectErrOut:   "",
			expectCommand:  "echo hi",
			seeOut:         "hi",
			seeErr:         "oops",
			seeOutWant:     true,
			seeErrWant:     false,
		},
		{
			name:           "failed result with stderr",
			res:            NewResult(2, "cmd", "out", "err msg"),
			expectSuccess:  false,
			expectFailed:   true,
			expectExitCode: 2,
			expectOutput:   "out",
			expectErrOut:   "err msg",
			expectCommand:  "cmd",
			seeOut:         "nope",
			seeErr:         "err",
			seeOutWant:     false,
			seeErrWant:     true,
		},
		{
			name:           "empty needle returns false",
			res:            NewResult(0, "cmd", "abc", "xyz"),
			expectSuccess:  true,
			expectFailed:   false,
			expectExitCode: 0,
			expectOutput:   "abc",
			expectErrOut:   "xyz",
			expectCommand:  "cmd",
			seeOut:         "",
			seeErr:         "",
			seeOutWant:     false,
			seeErrWant:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectSuccess, test.res.Successful())
			assert.Equal(t, test.expectFailed, test.res.Failed())
			assert.Equal(t, test.expectExitCode, test.res.ExitCode())
			assert.Equal(t, test.expectOutput, test.res.Output())
			assert.Equal(t, test.expectErrOut, test.res.ErrorOutput())
			assert.Equal(t, test.expectCommand, test.res.Command())
			assert.Equal(t, test.seeOutWant, test.res.SeeInOutput(test.seeOut))
			assert.Equal(t, test.seeErrWant, test.res.SeeInErrorOutput(test.seeErr))
		})
	}
}
