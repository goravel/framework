package console

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console"
	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestExecuteCommand(t *testing.T) {
	var (
		mockCtx *mocksconsole.Context
	)

	beforeEach := func() {
		mockCtx = mocksconsole.NewContext(t)
	}

	tests := []struct {
		name    string
		cmd     *exec.Cmd
		message []string
		setup   func()
		expect  error
	}{
		{
			name: "execute command failed",
			cmd:  exec.Command("unknown", "command"),
			setup: func() {
				mockCtx.EXPECT().Spinner("> @unknown command", mock.Anything).RunAndReturn(func(_ string, option console.SpinnerOption) error {
					return option.Action()
				}).Once()
			},
			expect: &exec.Error{
				Name: "unknown",
				Err:  exec.ErrNotFound,
			},
		},
		{
			name: "execute command success",
			cmd:  exec.Command("ls"),
			setup: func() {
				mockCtx.EXPECT().Spinner("> @ls", mock.Anything).RunAndReturn(func(_ string, option console.SpinnerOption) error {
					return option.Action()
				}).Once()
			},
			expect: nil,
		},
		{
			name:    "execute command with spinner message",
			cmd:     exec.Command("ls"),
			message: []string{"list files"},
			setup: func() {
				mockCtx.EXPECT().Spinner("list files", mock.Anything).RunAndReturn(func(_ string, option console.SpinnerOption) error {
					return option.Action()
				}).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()
			result := ExecuteCommand(mockCtx, tt.cmd, tt.message...)

			assert.Equal(t, tt.expect, result)
		})
	}
}
