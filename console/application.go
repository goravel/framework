package console

import (
	"context"
	"io"
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
)

type Application struct {
	commands   []cli.Command
	name       string
	usage      string
	usageText  string
	useArtisan bool
	version    string

	// For test
	writer io.Writer
}

// NewApplication Create a new Artisan application.
// Will add artisan flag to the command if useArtisan is true.
func NewApplication(name, usage, usageText, version string, useArtisan bool) console.Artisan {
	return &Application{
		name:       name,
		usage:      usage,
		usageText:  usageText,
		useArtisan: useArtisan,
		version:    version,
	}
}

func (r *Application) Register(commands []console.Command) {
	for _, item := range commands {
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

		r.commands = append(r.commands, cliCommand)
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
		if err := r.instance().Run(context.Background(), cliArgs); err != nil {
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

func (r *Application) instance() *cli.Command {
	command := &cli.Command{}
	command.CommandNotFound = commandNotFound
	// Create a copy of commands to avoid concurrent access issues
	command.Commands = r.copyCommands()
	command.Flags = []cli.Flag{noANSIFlag}
	command.Name = r.name
	command.OnUsageError = onUsageError
	command.Usage = r.usage
	command.UsageText = r.usageText
	command.Version = r.version
	command.Writer = r.writer
	command.HideHelp = true

	return command
}

// copyCommands creates a deep copy of the commands slice to prevent concurrent access issues
func (r *Application) copyCommands() []*cli.Command {
	if r.commands == nil {
		return nil
	}

	copied := make([]*cli.Command, len(r.commands))
	for i, cmd := range r.commands {
		// copied[i] = r.copyCommand(cmd)
		copied[i] = &cmd
	}
	return copied
}

// copyCommand creates a deep copy of a single CLI command
func (r *Application) copyCommand(original *cli.Command) *cli.Command {
	if original == nil {
		return nil
	}

	cmd := &cli.Command{
		Name:                   original.Name,
		Aliases:                append([]string(nil), original.Aliases...),
		Usage:                  original.Usage,
		UsageText:              original.UsageText,
		ArgsUsage:              original.ArgsUsage,
		Version:                original.Version,
		Description:            original.Description,
		Category:               original.Category,
		Hidden:                 original.Hidden,
		UseShortOptionHandling: original.UseShortOptionHandling,
		Action:                 original.Action,
		Before:                 original.Before,
		After:                  original.After,
		OnUsageError:           original.OnUsageError,
		Writer:                 original.Writer,
		ErrWriter:              original.ErrWriter,
		HideHelp:               original.HideHelp,
		HideHelpCommand:        original.HideHelpCommand,
		HideVersion:            original.HideVersion,
		CommandNotFound:        original.CommandNotFound,
		SkipFlagParsing:        original.SkipFlagParsing,
		AllowExtFlags:          original.AllowExtFlags,
	}

	// Copy flags
	if original.Flags != nil {
		cmd.Flags = make([]cli.Flag, len(original.Flags))
		copy(cmd.Flags, original.Flags)
	}

	// Copy subcommands recursively
	if original.Commands != nil {
		cmd.Commands = make([]*cli.Command, len(original.Commands))
		for i, subcmd := range original.Commands {
			cmd.Commands[i] = r.copyCommand(subcmd)
		}
	}

	// Copy mutually exclusive flags
	if original.MutuallyExclusiveFlags != nil {
		cmd.MutuallyExclusiveFlags = make([]cli.MutuallyExclusiveFlags, len(original.MutuallyExclusiveFlags))
		for i, mef := range original.MutuallyExclusiveFlags {
			cmd.MutuallyExclusiveFlags[i] = cli.MutuallyExclusiveFlags{
				Required: mef.Required,
				Category: mef.Category,
			}
			if mef.Flags != nil {
				cmd.MutuallyExclusiveFlags[i].Flags = make([][]cli.Flag, len(mef.Flags))
				for j, flagGroup := range mef.Flags {
					cmd.MutuallyExclusiveFlags[i].Flags[j] = make([]cli.Flag, len(flagGroup))
					copy(cmd.MutuallyExclusiveFlags[i].Flags[j], flagGroup)
				}
			}
		}
	}

	return cmd
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
