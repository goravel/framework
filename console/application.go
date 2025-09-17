package console

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
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
	instance   *cli.Command
	useArtisan bool
}

// NewApplication Create a new Artisan application.
// Will add artisan flag to the command if useArtisan is true.
func NewApplication(name, usage, usageText, version string, useArtisan bool) console.Artisan {
	instance := &cli.Command{}
	instance.Name = name
	instance.Usage = usage
	instance.UsageText = usageText
	instance.Version = version
	instance.CommandNotFound = commandNotFound
	instance.OnUsageError = onUsageError
	instance.Flags = []cli.Flag{noANSIFlag}

	return &Application{
		instance:   instance,
		useArtisan: useArtisan,
	}
}

func (r *Application) Register(commands []console.Command) {
	for _, item := range commands {
		item := item
		arguments, err := argumentsToCliArgs(item.Extend().Arguments)
		if err != nil {
			color.Errorln(fmt.Sprintf("Registration of command '%s' failed: %s", item.Signature(), err.Error()))
		}
		cliCommand := cli.Command{
			Name:  item.Signature(),
			Usage: item.Description(),
			Action: func(_ context.Context, cmd *cli.Command) error {
				return item.Handle(NewCliContext(cmd))
			},
			Category:     item.Extend().Category,
			ArgsUsage:    item.Extend().ArgsUsage,
			Flags:        flagsToCliFlags(item.Extend().Flags),
			Arguments:    arguments,
			OnUsageError: onUsageError,
		}
		r.instance.Commands = append(r.instance.Commands, &cliCommand)
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
		// Add --help if no command argument is provided.
		if artisanIndex+1 == len(args) {
			args = append(args, "--help")
		}

		cliArgs := append([]string{args[0]}, args[artisanIndex+1:]...)
		if err := r.instance.Run(context.Background(), cliArgs); err != nil {
			if exitIfArtisan {
				panic(err.Error())
			}

			return err
		}

		if exitIfArtisan {
			os.Exit(0)
		}
	}

	return nil
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
		if v.MinOccurrences() != 0 && !previousIsRequired {
			return nil, fmt.Errorf("required argument '%s' should be placed before any not-required arguments", v.ArgumentName())
		}
		if v.MinOccurrences() != 0 {
			previousIsRequired = true
		} else {
			previousIsRequired = false
		}
		switch arg := v.(type) {
		case *command.ArgumentFloat32:
			cliArgs = append(cliArgs, &cli.Float32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentFloat64:
			cliArgs = append(cliArgs, &cli.Float64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt:
			cliArgs = append(cliArgs, &cli.IntArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt8:
			cliArgs = append(cliArgs, &cli.Int8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt16:
			cliArgs = append(cliArgs, &cli.Int16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt32:
			cliArgs = append(cliArgs, &cli.Int32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt64:
			cliArgs = append(cliArgs, &cli.Int64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentString:
			cliArgs = append(cliArgs, &cli.StringArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentTimestamp:
			cliArgs = append(cliArgs, &cli.TimestampArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
				Config: cli.TimestampConfig{
					Layouts: []string{time.RFC3339},
				},
			})
		case *command.ArgumentUint:
			cliArgs = append(cliArgs, &cli.UintArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint8:
			cliArgs = append(cliArgs, &cli.Uint8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint16:
			cliArgs = append(cliArgs, &cli.Uint16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint32:
			cliArgs = append(cliArgs, &cli.Uint32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint64:
			cliArgs = append(cliArgs, &cli.Uint64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})

		case *command.ArgumentFloat32Slice:
			cliArgs = append(cliArgs, &cli.Float32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentFloat64Slice:
			cliArgs = append(cliArgs, &cli.Float64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentIntSlice:
			cliArgs = append(cliArgs, &cli.IntArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt8Slice:
			cliArgs = append(cliArgs, &cli.Int8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt16Slice:
			cliArgs = append(cliArgs, &cli.Int16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt32Slice:
			cliArgs = append(cliArgs, &cli.Int32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentInt64Slice:
			cliArgs = append(cliArgs, &cli.Int64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentStringSlice:
			cliArgs = append(cliArgs, &cli.StringArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentTimestampSlice:
			cliArgs = append(cliArgs, &cli.TimestampArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
				Config: cli.TimestampConfig{
					Layouts: []string{time.RFC3339},
				},
			})
		case *command.ArgumentUintSlice:
			cliArgs = append(cliArgs, &cli.UintArgs{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint8Slice:
			cliArgs = append(cliArgs, &cli.Uint8Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint16Slice:
			cliArgs = append(cliArgs, &cli.Uint16Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint32Slice:
			cliArgs = append(cliArgs, &cli.Uint32Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		case *command.ArgumentUint64Slice:
			cliArgs = append(cliArgs, &cli.Uint64Args{
				Name:      arg.Name,
				UsageText: arg.Usage,
				Value:     arg.Value,
				Min:       arg.MinOccurrences(),
				Max:       arg.MaxOccurrences(),
			})
		default:
			return nil, fmt.Errorf("unknown type of console command argument %T, with value %+v", arg, arg)
		}
	}
	return cliArgs, nil
}
