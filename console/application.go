package console

import (
	"context"
	"io"
	"maps"
	"os"
	"os/signal"
	stdpath "path"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
)

var (
	noANSI     bool
	noANSIFlag = &cli.BoolFlag{
		Name:        "no-ansi",
		Destination: &noANSI,
		HideDefault: true,
		Usage:       "Force disable ANSI output",
	}
)

type Application struct {
	commands   []console.Command
	hooks      map[string][]console.CommandHook
	name       string
	usage      string
	usageText  string
	useArtisan bool
	version    string
	writer     io.Writer
	ctx        context.Context
}

// NewApplication Create a new Artisan application.
// Will add artisan flag to the command if useArtisan is true.
// hooks run around matching commands, keyed by command signature or '*'
// pattern.
func NewApplication(name, usage, usageText, version string, useArtisan bool, hooks map[string][]console.CommandHook) *Application {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &Application{
		hooks:      hooks,
		name:       name,
		usage:      usage,
		usageText:  usageText,
		useArtisan: useArtisan,
		version:    version,
		writer:     os.Stdout,
		ctx:        ctx,
	}
}

// Call Run an Artisan console command by name.
func (r *Application) Call(command string) error {
	if len(os.Args) == 0 {
		return nil
	}

	commands := []string{os.Args[0]}

	if r.useArtisan {
		commands = append(commands, "artisan")
	}

	return r.Run(append(commands, strings.Split(command, " ")...), false)
}

// CallAndExit Run an Artisan console command by name and exit.
func (r *Application) CallAndExit(command string) {
	if len(os.Args) == 0 {
		return
	}

	commands := []string{os.Args[0]}

	if r.useArtisan {
		commands = append(commands, "artisan")
	}

	_ = r.Run(append(commands, strings.Split(command, " ")...), true)
}

// Register commands to the application.
func (r *Application) Register(commands []console.Command) {
	r.commands = append(r.commands, commands...)
}

// Run a command. Args come from os.Args.
func (r *Application) Run(args []string, exitIfArtisan bool) error {
	if noANSI || env.IsNoANSI() || slices.Contains(args, "--no-ansi") {
		color.Disable()
	} else {
		color.Enable()
	}

	artisanIndex := -1
	if r.useArtisan {
		for i, arg := range args {
			if arg == "artisan" {
				artisanIndex = i
				break
			}
		}
	} else {
		artisanIndex = 0
	}

	if artisanIndex != -1 {
		command, err := r.command()
		if err != nil {
			return err
		}

		if artisanIndex+1 == len(args) {
			args = append(args, "list")
		}

		cliArgs := append([]string{args[0]}, args[artisanIndex+1:]...)
		if err := command.Run(r.ctx, cliArgs); err != nil {
			if exitIfArtisan {
				if !errors.Is(err, context.Canceled) {
					color.Errorln(err.Error())
					os.Exit(1)
				}
				os.Exit(0)
			}

			return err
		}

		if exitIfArtisan {
			os.Exit(0)
		}
	}

	return nil
}

// SetCommands Set the commands for the application.
func (r *Application) SetCommands(commands []console.Command) {
	r.commands = commands
}

func (r *Application) command() (*cli.Command, error) {
	cliCommands, err := commandsToCliCommands(r.commands, r.hooks)
	if err != nil {
		return nil, err
	}

	command := &cli.Command{}
	command.CommandNotFound = commandNotFound
	command.Commands = cliCommands
	command.Flags = []cli.Flag{noANSIFlag}
	command.Name = r.name
	command.OnUsageError = onUsageError
	command.Usage = r.usage
	command.UsageText = r.usageText
	command.Version = r.version
	command.Writer = r.writer

	// There is a concurrency issue with urfave/cli v3 when help is not hidden.
	command.HideHelp = true

	return command, nil
}

// hookDecisionTimeout bounds how long a cancelled Shutdownable command waits
// for its Before hooks to decide whether Handle runs. A hook that ignores
// cancellation (a prompt, blocking I/O) would otherwise hold Ctrl-C forever,
// so after this long Shutdown runs anyway, as it did before hooks existed.
var hookDecisionTimeout = 5 * time.Second

func shutdownCommand(shutdownable console.Shutdownable, cmd *cli.Command, arguments []command.Argument) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	return shutdownable.Shutdown(NewCliContext(shutdownCtx, cmd, arguments))
}

// FilterCommandsByAllowlist returns the subset of commands whose Signature()
// matches at least one entry in allowlist.
//
// Matching is signature-only (category never consulted):
//   - Entries without '*' are exact-matched against command.Signature().
//   - Entries containing '*' use glob matching via path.Match against
//     command.Signature() (only '*' triggers the glob path; '?' is matched
//     literally as an exact match).
//
// Special cases:
//   - allowlist == nil            → every command is kept (no filter applied).
//   - allowlist resolves to no usable entries → every command is dropped.
func FilterCommandsByAllowlist(commands []console.Command, allowlist []string) []console.Command {
	if allowlist == nil {
		return commands
	}
	exact := make(map[string]struct{}, len(allowlist))
	var globs []string
	for _, s := range allowlist {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.Contains(s, "*") {
			globs = append(globs, s)
		} else {
			exact[s] = struct{}{}
		}
	}
	if len(exact) == 0 && len(globs) == 0 {
		return nil
	}
	kept := make([]console.Command, 0, len(commands))
	for _, cmd := range commands {
		sig := cmd.Signature()
		if _, ok := exact[sig]; ok {
			kept = append(kept, cmd)
			continue
		}
		for _, pattern := range globs {
			if ok, _ := stdpath.Match(pattern, sig); ok {
				kept = append(kept, cmd)
				break
			}
		}
	}
	return kept
}

// hooksForCommand returns the hooks that apply to item, outermost first: the
// matching keys in sorted order, then the hooks of each key in slice order.
func hooksForCommand(item console.Command, hooks map[string][]console.CommandHook) []console.CommandHook {
	sig := item.Signature()
	var matched []console.CommandHook
	for _, pattern := range slices.Sorted(maps.Keys(hooks)) {
		if !signatureMatches(pattern, sig) {
			continue
		}
		for _, hook := range hooks[pattern] {
			if hook != nil {
				matched = append(matched, hook)
			}
		}
	}

	return matched
}

// commandFlags returns the CLI flags for item: its own flags, then the flags of
// each matching hook in the order hooksForCommand returns them, then the
// default help and no-ansi flags. A hook flag whose name or alias is already taken, or is reserved for
// a default flag, is an error.
func commandFlags(item console.Command, hooks []console.CommandHook) ([]cli.Flag, error) {
	flags := convertFlags(item.Extend().Flags)

	taken := map[string]bool{"help": true, "h": true, noANSIFlag.Name: true}
	for _, flag := range flags {
		for _, name := range flag.Names() {
			taken[name] = true
		}
	}

	for _, hook := range hooks {
		for _, flag := range convertFlags(hook.Flags()) {
			names := flag.Names()
			for _, name := range names {
				if taken[name] {
					return nil, errors.ConsoleCommandHookFlagConflict.Args(name)
				}
			}
			for _, name := range names {
				taken[name] = true
			}
			flags = append(flags, flag)
		}
	}

	return withDefaultFlags(flags), nil
}

// signatureMatches reports whether pattern matches sig, using the same rules as
// FilterCommandsByAllowlist: exact match, or path.Match when pattern contains '*'.
func signatureMatches(pattern, sig string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	if strings.Contains(pattern, "*") {
		ok, err := stdpath.Match(pattern, sig)
		return err == nil && ok
	}

	return pattern == sig
}

// handleWithHooks runs the Before hooks in order, then Handle, then the After
// hooks in reverse order. A Before that returns false or an error skips Handle
// and every After. started is closed just before Handle runs. A Handle panic
// becomes an error so the After hooks still run and see it.
func handleWithHooks(item console.Command, hooks []console.CommandHook, ctx console.Context, started chan<- struct{}) (handled bool, err error) {
	for _, hook := range hooks {
		proceed, err := hook.Before(ctx)
		if err != nil || !proceed {
			return false, err
		}
	}

	close(started)
	err = handleRecovered(item, ctx)

	for i := len(hooks) - 1; i >= 0; i-- {
		err = hooks[i].After(ctx, err)
	}

	return true, err
}

// handleRecovered calls Handle, returning a panic as an error.
func handleRecovered(item console.Command, ctx console.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.ConsoleCommandPanicInHandle.Args(r)
		}
	}()

	return item.Handle(ctx)
}

func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

type handleResult struct {
	err     error
	handled bool
}

func commandsToCliCommands(commands []console.Command, hooks map[string][]console.CommandHook) ([]*cli.Command, error) {
	cliCommands := make([]*cli.Command, len(commands))

	for i, item := range commands {
		arguments := item.Extend().Arguments
		cliArguments, err := argumentsToCliArgs(arguments)
		if err != nil {
			return nil, errors.ConsoleCommandRegisterFailed.Args(item.Signature(), err)
		}
		itemHooks := hooksForCommand(item, hooks)
		flags, err := commandFlags(item, itemHooks)
		if err != nil {
			return nil, errors.ConsoleCommandRegisterFailed.Args(item.Signature(), err)
		}
		cliCommands[i] = &cli.Command{
			Name:  item.Signature(),
			Usage: item.Description(),
			Action: func(ctx context.Context, cmd *cli.Command) error {
				cliCtx := NewCliContext(ctx, cmd, arguments)
				if cliCtx.OptionBool("help") {
					return cli.ShowCommandHelp(ctx, cmd, cmd.Name)
				}

				started := make(chan struct{})
				resultCh := make(chan handleResult, 1)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							resultCh <- handleResult{err: errors.ConsoleCommandPanicInHandle.Args(r), handled: isClosed(started)}
						}
					}()
					handled, err := handleWithHooks(item, itemHooks, cliCtx, started)
					resultCh <- handleResult{err: err, handled: handled}
				}()

				shutdownable, ok := item.(console.Shutdownable)

				if ok {
					select {
					case result := <-resultCh:
						// Nothing to clean up when a Before hook skipped Handle.
						if !result.handled {
							return result.err
						}
						if err := shutdownCommand(shutdownable, cmd, arguments); err != nil {
							color.Errorln("shutdown error:", err.Error())
						}
						return result.err
					case <-ctx.Done():
						// Wait for the Before hooks to decide, so Shutdown only
						// runs for a Handle that was actually started. The wait
						// is bounded: a hook that never observes cancellation
						// must not block Ctrl-C.
						select {
						case <-started:
						case result := <-resultCh:
							if !result.handled {
								if result.err != nil {
									return result.err
								}
								return ctx.Err()
							}
						case <-time.After(hookDecisionTimeout):
						}
						return shutdownCommand(shutdownable, cmd, arguments)
					}
				}

				select {
				case result := <-resultCh:
					return result.err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			Category:     item.Extend().Category,
			ArgsUsage:    item.Extend().ArgsUsage,
			Flags:        flags,
			Arguments:    cliArguments,
			OnUsageError: onUsageError,
		}
	}

	return cliCommands, nil
}

// convertFlags converts command flags to CLI flags, without the defaults.
func convertFlags(flags []command.Flag) []cli.Flag {
	var cliFlags []cli.Flag
	for _, flag := range flags {
		switch flag.Type() {
		case command.FlagTypeBool:
			flag := flag.(*command.BoolFlag)
			cliFlags = append(cliFlags, &cli.BoolFlag{
				Name:        flag.Name,
				Aliases:     flag.Aliases,
				HideDefault: flag.DisableDefaultText,
				Usage:       flag.Usage,
				Required:    flag.Required,
				Value:       flag.Value,
			})
		case command.FlagTypeFloat64:
			flag := flag.(*command.Float64Flag)
			cliFlags = append(cliFlags, &cli.FloatFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeFloat64Slice:
			flag := flag.(*command.Float64SliceFlag)
			cliFlags = append(cliFlags, &cli.FloatSliceFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    cli.NewFloatSlice(flag.Value...).Value(),
			})
		case command.FlagTypeInt:
			flag := flag.(*command.IntFlag)
			cliFlags = append(cliFlags, &cli.IntFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeIntSlice:
			flag := flag.(*command.IntSliceFlag)
			cliFlags = append(cliFlags, &cli.IntSliceFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeInt64:
			flag := flag.(*command.Int64Flag)
			cliFlags = append(cliFlags, &cli.Int64Flag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeInt64Slice:
			flag := flag.(*command.Int64SliceFlag)
			cliFlags = append(cliFlags, &cli.Int64SliceFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeString:
			flag := flag.(*command.StringFlag)
			cliFlags = append(cliFlags, &cli.StringFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeStringSlice:
			flag := flag.(*command.StringSliceFlag)
			cliFlags = append(cliFlags, &cli.StringSliceFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		}
	}

	return cliFlags
}

// withDefaultFlags appends the help flag (unless already defined) and the
// no-ansi flag.
func withDefaultFlags(cliFlags []cli.Flag) []cli.Flag {
	var (
		existHelp bool
		existH    bool
	)
	for _, flag := range cliFlags {
		names := flag.Names()
		if slices.Contains(names, "help") {
			existHelp = true
		}
		if slices.Contains(names, "h") {
			existH = true
		}
	}

	if !existHelp {
		helpFlag := &cli.BoolFlag{
			Name:        "help",
			Usage:       "Show help",
			HideDefault: true,
		}
		if !existH {
			helpFlag.Aliases = []string{"h"}
		}
		cliFlags = append(cliFlags, helpFlag)
	}

	cliFlags = append(cliFlags, noANSIFlag)

	return cliFlags
}

func argumentsToCliArgs(args []command.Argument) ([]cli.Argument, error) {
	len := len(args)
	if len == 0 {
		return nil, nil
	}
	cliArgs := make([]cli.Argument, 0, len)
	previousIsRequired := true
	for _, v := range args {
		if v.GetMin() != 0 && !previousIsRequired {
			return nil, errors.ConsoleCommandRequiredArgumentWrongOrder.Args(v.GetName())
		}
		if v.GetMin() != 0 {
			previousIsRequired = true
		} else {
			previousIsRequired = false
		}
		switch arg := v.(type) {
		case *command.ArgumentFloat32:
			cliArgs = append(cliArgs, &cli.Float32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentFloat64:
			cliArgs = append(cliArgs, &cli.Float64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt:
			cliArgs = append(cliArgs, &cli.IntArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt8:
			cliArgs = append(cliArgs, &cli.Int8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt16:
			cliArgs = append(cliArgs, &cli.Int16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt32:
			cliArgs = append(cliArgs, &cli.Int32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt64:
			cliArgs = append(cliArgs, &cli.Int64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentString:
			cliArgs = append(cliArgs, &cli.StringArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentTimestamp:
			cliArgs = append(cliArgs, &cli.TimestampArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
				Config: cli.TimestampConfig{
					Layouts: arg.Layouts,
				},
			})
		case *command.ArgumentUint:
			cliArgs = append(cliArgs, &cli.UintArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint8:
			cliArgs = append(cliArgs, &cli.Uint8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint16:
			cliArgs = append(cliArgs, &cli.Uint16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint32:
			cliArgs = append(cliArgs, &cli.Uint32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint64:
			cliArgs = append(cliArgs, &cli.Uint64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})

		case *command.ArgumentFloat32Slice:
			cliArgs = append(cliArgs, &cli.Float32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentFloat64Slice:
			cliArgs = append(cliArgs, &cli.Float64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentIntSlice:
			cliArgs = append(cliArgs, &cli.IntArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt8Slice:
			cliArgs = append(cliArgs, &cli.Int8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt16Slice:
			cliArgs = append(cliArgs, &cli.Int16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt32Slice:
			cliArgs = append(cliArgs, &cli.Int32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentInt64Slice:
			cliArgs = append(cliArgs, &cli.Int64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentStringSlice:
			cliArgs = append(cliArgs, &cli.StringArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentTimestampSlice:
			cliArgs = append(cliArgs, &cli.TimestampArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
				Config: cli.TimestampConfig{
					Layouts: arg.Layouts,
				},
			})
		case *command.ArgumentUintSlice:
			cliArgs = append(cliArgs, &cli.UintArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint8Slice:
			cliArgs = append(cliArgs, &cli.Uint8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint16Slice:
			cliArgs = append(cliArgs, &cli.Uint16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint32Slice:
			cliArgs = append(cliArgs, &cli.Uint32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		case *command.ArgumentUint64Slice:
			cliArgs = append(cliArgs, &cli.Uint64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Min:       arg.GetMin(),
				Max:       arg.GetMax(),
			})
		default:
			return nil, errors.ConsoleCommandArgumentUnknownType.Args(arg, arg)
		}
	}
	return cliArgs, nil
}
