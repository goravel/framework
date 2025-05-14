package console

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/goravel/framework/contracts/console"
)

func ExecuteCommand(ctx console.Context, cmd *exec.Cmd, message ...string) error {
	if len(message) == 0 {
		message = []string{fmt.Sprintf("> @%s", strings.Join(cmd.Args, " "))}
	}

	return ctx.Spinner(message[0], console.SpinnerOption{
		Action: func() error {
			output, err := cmd.CombinedOutput()
			if err != nil && len(output) > 0 {
				err = errors.New(strings.TrimSpace(strings.ReplaceAll(string(output), err.Error(), "")))
			}

			return err
		},
	})
}
