package process

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestResult_Methods(t *testing.T) {
	err := errors.New("hello")
	res := NewResult(err, 0, "cmd arg1", "out", "err")
	assert.ErrorIs(t, err, res.Error())
	assert.True(t, res.Successful())
	assert.False(t, res.Failed())
	assert.Equal(t, 0, res.ExitCode())
	assert.Equal(t, "out", res.Output())
	assert.Equal(t, "err", res.ErrorOutput())
	assert.Equal(t, "cmd arg1", res.Command())
	assert.True(t, res.SeeInOutput("ou"))
	assert.True(t, res.SeeInErrorOutput("er"))
}

func TestResult_NilReceiverSafety(t *testing.T) {
	var res *Result
	assert.False(t, res.Successful())
	assert.True(t, res.Failed())
	assert.Equal(t, -1, res.ExitCode())
	assert.Equal(t, "", res.Output())
	assert.Equal(t, "", res.ErrorOutput())
	assert.Equal(t, "", res.Command())
	assert.False(t, res.SeeInOutput("x"))
	assert.False(t, res.SeeInErrorOutput("y"))
}

func TestResultMethods_TableDriven(t *testing.T) {
	dummyErr := errors.New("error")

	tests := []struct {
		name           string
		res            *Result
		expectSuccess  bool
		expectFailed   bool
		expectExitCode int
		expectOutput   string
		expectErrOut   string
		expectedErr    error
		expectCommand  string
		seeOut         string
		seeErr         string
		seeOutWant     bool
		seeErrOutWant  bool
	}{
		{
			name:           "nil receiver",
			res:            nil,
			expectSuccess:  false,
			expectFailed:   true,
			expectExitCode: -1,
			expectOutput:   "",
			expectErrOut:   "",
			expectedErr:    nil,
			expectCommand:  "",
			seeOut:         "anything",
			seeErr:         "anything",
			seeOutWant:     false,
			seeErrOutWant:  false,
		},
		{
			name:           "successful result",
			res:            NewResult(nil, 0, "echo hi", "hi", ""),
			expectSuccess:  true,
			expectFailed:   false,
			expectExitCode: 0,
			expectOutput:   "hi",
			expectErrOut:   "",
			expectedErr:    nil,
			expectCommand:  "echo hi",
			seeOut:         "hi",
			seeErr:         "oops",
			seeOutWant:     true,
			seeErrOutWant:  false,
		},
		{
			name:           "failed result with stderr",
			res:            NewResult(nil, 2, "cmd", "out", "err msg"),
			expectSuccess:  false,
			expectFailed:   true,
			expectExitCode: 2,
			expectOutput:   "out",
			expectErrOut:   "err msg",
			expectCommand:  "cmd",
			seeOut:         "nope",
			seeErr:         "err",
			seeOutWant:     false,
			seeErrOutWant:  true,
		},
		{
			name:           "empty needle returns false",
			res:            NewResult(nil, 0, "cmd", "abc", "xyz"),
			expectSuccess:  true,
			expectFailed:   false,
			expectExitCode: 0,
			expectOutput:   "abc",
			expectErrOut:   "xyz",
			expectCommand:  "cmd",
			seeOut:         "",
			seeErr:         "",
			seeOutWant:     false,
			seeErrOutWant:  false,
		},
		{
			name:           "command wait error",
			res:            NewResult(dummyErr, 0, "cmd", "abc", "xyz"),
			expectSuccess:  true,
			expectFailed:   false,
			expectExitCode: 0,
			expectOutput:   "abc",
			expectErrOut:   "xyz",
			expectCommand:  "cmd",
			expectedErr:    dummyErr,
			seeOut:         "",
			seeErr:         "",
			seeOutWant:     false,
			seeErrOutWant:  false,
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
			assert.ErrorIs(t, test.expectedErr, test.res.Error())
			assert.Equal(t, test.seeOutWant, test.res.SeeInOutput(test.seeOut))
			assert.Equal(t, test.seeErrOutWant, test.res.SeeInErrorOutput(test.seeErr))
		})
	}
}
