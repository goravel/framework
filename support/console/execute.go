package console

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/goravel/framework/contracts/console"
)

func ExecuteCommand(ctx console.Context, cmd *exec.Cmd) error {
	return ctx.Spinner(fmt.Sprintf("> @%s", strings.Join(cmd.Args, " ")), console.SpinnerOption{
		Action: func() error {
			output, err := cmd.CombinedOutput()
			if err != nil && len(output) > 0 {
				err = errors.New(strings.TrimSpace(strings.ReplaceAll(string(output), err.Error(), "")))
			}

			return err
		},
	})
}
