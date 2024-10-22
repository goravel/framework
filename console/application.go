package console

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type Application struct {
	instance  *cli.App
	isArtisan bool
}

func NewApplication(name, usage, usageText, version string, artisan ...bool) console.Artisan {
	instance := cli.NewApp()
	instance.Name = name
	instance.Usage = usage
	instance.UsageText = usageText
	instance.Version = version
	isArtisan := len(artisan) > 0 && artisan[0]

	return &Application{
		instance:  instance,
		isArtisan: isArtisan,
	}
}

func (r *Application) Register(commands []console.Command) {
	for _, item := range commands {
		item := item
		cliCommand := cli.Command{
			Name:  item.Signature(),
			Usage: item.Description(),
			Action: func(ctx *cli.Context) error {
				return item.Handle(NewCliContext(ctx))
			},
			Category: item.Extend().Category,
			Flags:    flagsToCliFlags(item.Extend().Flags),
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

	if r.isArtisan {
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

	if r.isArtisan {
		commands = append(commands, "artisan")
	}

	_ = r.Run(append(commands, strings.Split(command, " ")...), true)
}

// Run a command. Args come from os.Args.
func (r *Application) Run(args []string, exitIfArtisan bool) error {
	artisanIndex := -1
	if r.isArtisan {
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
		if err := r.instance.Run(cliArgs); err != nil {
			if exitIfArtisan {
				panic(err.Error())
			}

			return err
		}

		if exitIfArtisan {
			os.Exit(1)
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
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeFloat64:
			flag := flag.(*command.Float64Flag)
			cliFlags = append(cliFlags, &cli.Float64Flag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			})
		case command.FlagTypeFloat64Slice:
			flag := flag.(*command.Float64SliceFlag)
			cliFlags = append(cliFlags, &cli.Float64SliceFlag{
				Name:     flag.Name,
				Aliases:  flag.Aliases,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    cli.NewFloat64Slice(flag.Value...),
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
				Value:    cli.NewIntSlice(flag.Value...),
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
				Value:    cli.NewInt64Slice(flag.Value...),
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
				Value:    cli.NewStringSlice(flag.Value...),
			})
		}
	}

	return cliFlags
}
