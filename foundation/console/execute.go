package console

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/goravel/framework/contracts/console"
)

func execute(ctx console.Context, cmd *exec.Cmd) (err error, msg string) {
	var output []byte
	if err = ctx.Spinner(fmt.Sprintf("> @%s", strings.Join(cmd.Args, " ")), console.SpinnerOption{
		Action: func() error {
			output, err = cmd.CombinedOutput()

			return err
		},
	}); err != nil {
		msg = strings.TrimSpace(strings.ReplaceAll(string(output), err.Error(), ""))
	}

	return
}
