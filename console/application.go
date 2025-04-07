package console

import (
	"context"
	"os"
	"slices"
	"strings"

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

	globalFlags = []cli.Flag{
		noANSIFlag,
		cli.HelpFlag,
		cli.VersionFlag,
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
		cliCommand := cli.Command{
			Name:  item.Signature(),
			Usage: item.Description(),
			Action: func(_ context.Context, cmd *cli.Command) error {
				return item.Handle(NewCliContext(cmd))
			},
			Category:     item.Extend().Category,
			ArgsUsage:    item.Extend().ArgsUsage,
			Flags:        flagsToCliFlags(item.Extend().Flags),
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
				Value:    int64(flag.Value),
			})
		case command.FlagTypeIntSlice:
			flag := flag.(*command.IntSliceFlag)
			var int64Slice []int64
			for _, v := range flag.Value {
				int64Slice = append(int64Slice, int64(v))
			}

			cliFlags = append(cliFlags, &cli.IntSliceFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    cli.NewIntSlice(int64Slice...).Value(),
			})
		case command.FlagTypeInt64:
			flag := flag.(*command.Int64Flag)
			cliFlags = append(cliFlags, &cli.IntFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeInt64Slice:
			flag := flag.(*command.Int64SliceFlag)
			cliFlags = append(cliFlags, &cli.IntSliceFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    cli.NewIntSlice(flag.Value...).Value(),
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
				Value:    cli.NewStringSlice(flag.Value...).Value(),
			})
		}
	}

	return cliFlags
}
