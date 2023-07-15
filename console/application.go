package console

import (
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/urfave/cli/v2"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support"
)

type Application struct {
	instance *cli.App
}

func NewApplication() console.Artisan {
	instance := cli.NewApp()
	instance.Name = "Goravel Framework"
	instance.Usage = support.Version
	instance.UsageText = "artisan [global options] command [options] [arguments...]"

	return &Application{instance}
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
		}

		cliCommand.Category = item.Extend().Category
		cliCommand.Flags = flagsToCliFlags(item.Extend().Flags)
		c.instance.Commands = append(c.instance.Commands, &cliCommand)
	}
}

// Call Run an Artisan console command by name.
func (c *Application) Call(command string) {
	c.Run(append([]string{os.Args[0], "artisan"}, strings.Split(command, " ")...), false)
}

// CallAndExit Run an Artisan console command by name and exit.
func (c *Application) CallAndExit(command string) {
	c.Run(append([]string{os.Args[0], "artisan"}, strings.Split(command, " ")...), true)
}

// Run a command. Args come from os.Args.
func (c *Application) Run(args []string, exitIfArtisan bool) {
	artisanIndex := -1
	for i, arg := range args {
		if arg == "artisan" {
			artisanIndex = i
			break
		}
	}

	if artisanIndex != -1 {
		// Add --help if no command argument is provided.
		if artisanIndex+1 == len(args) {
			args = append(args, "--help")
		}

		if args[artisanIndex+1] != "-V" && args[artisanIndex+1] != "--version" {
			cliArgs := append([]string{args[0]}, args[artisanIndex+1:]...)
			if err := c.instance.Run(cliArgs); err != nil {
				panic(err.Error())
			}
		}

		printResult(args[artisanIndex+1])

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

func printResult(command string) {
	switch command {
	case "make:command":
		color.Greenln("Console command created successfully")
	case "-V", "--version":
		color.Greenln("Goravel Framework " + support.Version)
	}
}
