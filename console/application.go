package console

import (
	"context"
	"io"
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
func NewApplication(name, usage, usageText, version string, useArtisan bool) *Application {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &Application{
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
	cliCommands, err := commandsToCliCommands(r.commands)
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

func shutdownCommand(shutdownable console.Shutdownable, cmd *cli.Command, arguments []command.Argument) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	return shutdownable.Shutdown(NewCliContext(shutdownCtx, cmd, arguments))
}

// FilterCommandsByAllowlist returns the subset of commands whose Signature()
// matches at least one entry in allowlist. Each entry is matched in one of
// two ways:
//
//   - Exact match (no wildcard) — checked against command.Signature().
//   - Glob match (the entry contains '*') — checked against
//     command.Signature() using stdpath.Match. '*' matches any sequence of
//     non-'/' characters. '?' is not a wildcard.
//
// Category is never consulted. The filter is signature-only.
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

func commandsToCliCommands(commands []console.Command) ([]*cli.Command, error) {
	cliCommands := make([]*cli.Command, len(commands))

	for i, item := range commands {
		arguments := item.Extend().Arguments
		cliArguments, err := argumentsToCliArgs(arguments)
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

				errCh := make(chan error, 1)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							errCh <- errors.ConsoleCommandPanicInHandle.Args(r)
						}
					}()
					errCh <- item.Handle(cliCtx)
				}()

				shutdownable, ok := item.(console.Shutdownable)

				if ok {
					select {
					case handleErr := <-errCh:
						if err := shutdownCommand(shutdownable, cmd, arguments); err != nil {
							color.Errorln("shutdown error:", err.Error())
						}
						return handleErr
					case <-ctx.Done():
						return shutdownCommand(shutdownable, cmd, arguments)
					}
				}

				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			Category:     item.Extend().Category,
			ArgsUsage:    item.Extend().ArgsUsage,
			Flags:        flagsToCliFlags(item.Extend().Flags),
			Arguments:    cliArguments,
			OnUsageError: onUsageError,
		}
	}

	return cliCommands, nil
}

func flagsToCliFlags(flags []command.Flag) []cli.Flag {
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
