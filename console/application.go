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

func (c *Application) Register(commands []console.Command) {
	for _, item := range commands {
		item := item
		cliCommand := cli.Command{
			Name:  item.Signature(),
			Usage: item.Description(),
			Action: func(ctx *cli.Context) error {
				return item.Handle(&CliContext{ctx})
			},
			Category: item.Extend().Category,
			Flags:    flagsToCliFlags(item.Extend().Flags),
		}
		c.instance.Commands = append(c.instance.Commands, &cliCommand)
	}
}

// Call Run an Artisan console command by name.
func (c *Application) Call(command string) {
	commands := []string{os.Args[0]}
	if c.isArtisan {
		commands = append(commands, "artisan")
	}
	c.Run(append(commands, strings.Split(command, " ")...), false)
}

// CallAndExit Run an Artisan console command by name and exit.
func (c *Application) CallAndExit(command string) {
	commands := []string{os.Args[0]}
	if c.isArtisan {
		commands = append(commands, "artisan")
	}
	c.Run(append(commands, strings.Split(command, " ")...), true)
}

// Run a command. Args come from os.Args.
func (c *Application) Run(args []string, exitIfArtisan bool) {
	artisanIndex := -1
	if c.isArtisan {
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
		if err := c.instance.Run(cliArgs); err != nil {
			panic(err.Error())
		}

		if exitIfArtisan {
			os.Exit(0)
		}
	}
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
