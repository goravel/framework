package console

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v2"
	"github.com/xrash/smetrics"

	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
)

func init() {
	cli.HelpPrinterCustom = printHelpCustom
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.VersionPrinter = printVersion
	huh.ErrUserAborted = cli.Exit(color.Red().Sprint("Cancelled."), 0)
}

const maxLineLength = 10000

// Template for the help message.
var (
	appHelpTemplate = `{{$v := offset .Usage 6}}{{wrap (colorize .Usage) 3}}{{if .Version}} {{green (wrap .Version $v)}}{{end}}

{{ yellow "Usage:" }}
   {{if .UsageText}}{{wrap (colorize .UsageText) 3}}{{end}}{{if .VisibleFlags}}

{{ yellow "Global options:" }}{{template "flagTemplate" .}}{{end}}{{if .VisibleCommands}}

{{ yellow "Available commands:" }}{{template "commandTemplate" .}}{{end}}
`

	commandHelpTemplate = `{{ yellow "Description:" }}
   {{ (colorize .Usage) }}

{{ yellow "Usage:" }}
   {{template "usageTemplate" .}}{{if .VisibleFlags}}

{{ yellow "Options:" }}{{template "flagTemplate" .}}{{end}}
`
	commandTemplate = `{{ $cv := offsetCommands .VisibleCommands 5}}{{range .VisibleCategories}}{{if .Name}}
 {{yellow .Name}}:{{end}}{{range (sortCommands .VisibleCommands)}}
  {{$s := join .Names ", "}}{{green $s}}{{ $sp := subtract $cv (offset $s 3) }}{{ indent $sp ""}}{{wrap (colorize .Usage) $cv}}{{end}}{{end}}`
	flagTemplate = `{{ $cv := offsetFlags .VisibleFlags 5}}{{range  (sortFlags .VisibleFlags)}}
   {{$s := getFlagName .}}{{green $s}}{{ $sp := subtract $cv (offset $s 1) }}{{ indent $sp ""}}{{$us := (capitalize .Usage)}}{{wrap (colorize $us) $cv}}{{$df := getFlagDefaultText . }}{{if $df}} {{yellow $df}}{{end}}{{end}}`
	usageTemplate = `{{if .UsageText}}{{wrap (colorize .UsageText) 3}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}}{{if .ArgsUsage}}{{.ArgsUsage}}{{else}}{{if .Args}} [arguments...]{{end}}{{end}}{{end}}`
)

// colorsFuncMap is a map of functions for coloring text.
var colorsFuncMap = template.FuncMap{
	"black":   color.Black().Sprint,
	"blue":    color.Blue().Sprint,
	"cyan":    color.Cyan().Sprint,
	"default": color.Default().Sprint,
	"gray":    color.Gray().Sprint,
	"green":   color.Green().Sprint,
	"magenta": color.Magenta().Sprint,
	"red":     color.Red().Sprint,
	"white":   color.White().Sprint,
	"yellow":  color.Yellow().Sprint,
}

func capitalize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// colorize wraps the text in the default color
// support style tags like <fg=red>text</>
// more details in https://gookit.github.io/color/#/?id=tag-attributes
func colorize(text string) string {
	return color.Default().Sprint(text)

}

func commandNotFound(ctx *cli.Context, command string) {
	var (
		msgTxt     = fmt.Sprintf("Command '%s' is not defined.", command)
		suggestion string
	)
	if alternatives := findAlternatives(command, func() (collection []string) {
		for i := range ctx.App.Commands {
			collection = append(collection, ctx.App.Commands[i].Names()...)
		}
		return
	}()); len(alternatives) > 0 {
		if len(alternatives) == 1 {
			msgTxt = msgTxt + " Did you mean this?"
		} else {
			msgTxt = msgTxt + " Did you mean one of these?"
		}
		suggestion = "\n  " + strings.Join(alternatives, "\n  ")
	}
	color.Errorln(msgTxt)
	color.Gray().Println(suggestion)
}

func findAlternatives(name string, collection []string) (result []string) {
	var (
		threshold       = 1e3
		alternatives    = make(map[string]float64)
		collectionParts = make(map[string][]string)
	)
	for i := range collection {
		collectionParts[collection[i]] = strings.Split(collection[i], ":")
	}
	for i, sub := range strings.Split(name, ":") {
		for collectionName, parts := range collectionParts {
			exists := alternatives[collectionName] != 0
			if len(parts) <= i {
				if exists {
					alternatives[collectionName] += threshold
				}
				continue
			}
			lev := smetrics.WagnerFischer(sub, parts[i], 1, 1, 1)
			if float64(lev) <= float64(len(sub))/3 || strings.Contains(parts[i], sub) {
				if exists {
					alternatives[collectionName] += float64(lev)
				} else {
					alternatives[collectionName] = float64(lev)
				}
			} else if exists {
				alternatives[collectionName] += threshold
			}
		}
	}
	for _, item := range collection {
		lev := smetrics.WagnerFischer(name, item, 1, 1, 1)
		if float64(lev) <= float64(len(name))/3 || strings.Contains(item, name) {
			if alternatives[item] != 0 {
				alternatives[item] -= float64(lev)
			} else {
				alternatives[item] = float64(lev)
			}
		}
	}
	type scoredItem struct {
		name  string
		score float64
	}
	var sortedAlternatives []scoredItem
	for item, score := range alternatives {
		if score < 2*threshold {
			sortedAlternatives = append(sortedAlternatives, scoredItem{item, score})
		}
	}
	sort.Slice(sortedAlternatives, func(i, j int) bool {
		if sortedAlternatives[i].score == sortedAlternatives[j].score {
			return sortedAlternatives[i].name < sortedAlternatives[j].name
		}
		return sortedAlternatives[i].score < sortedAlternatives[j].score
	})
	for _, item := range sortedAlternatives {
		result = append(result, item.name)
	}
	return result
}

func getFlagDefaultText(flag cli.DocGenerationFlag) string {
	defaultValueString := ""
	if bf, ok := flag.(*cli.BoolFlag); !ok || !bf.DisableDefaultText {
		if s := flag.GetDefaultText(); s != "" {
			defaultValueString = fmt.Sprintf(`[default: %s]`, s)
		}
	}
	return defaultValueString
}

func getFlagName(flag cli.DocGenerationFlag) string {
	names := flag.Names()
	sort.Slice(names, func(i, j int) bool {
		return len(names[i]) < len(names[j])
	})
	prefixed := cli.FlagNamePrefixer(names, "")
	// If there is no short name, add some padding to align flag name.
	if len(names) == 1 {
		prefixed = "    " + prefixed
	}

	return prefixed
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}

func offset(input string, fixed int) int {
	return len(input) + fixed
}

func offsetCommands(cmd []*cli.Command, fixed int) int {
	var maxLen = 0
	for i := range cmd {
		if s := strings.Join(cmd[i].Names(), ", "); len(s) > maxLen {
			maxLen = len(s)
		}
	}
	return maxLen + fixed
}

func offsetFlags(flags []cli.Flag, fixed int) int {
	var maxLen = 0
	for i := range flags {
		if s := cli.FlagNamePrefixer(flags[i].Names(), ""); len(s) > maxLen {
			maxLen = len(s)
		}
	}
	return maxLen + fixed
}

func onUsageError(_ *cli.Context, err error, _ bool) error {
	if flag, ok := strings.CutPrefix(err.Error(), "flag provided but not defined: -"); ok {
		color.Red().Printfln("The '%s' option does not exist.", flag)
		return nil
	}
	if flag, ok := strings.CutPrefix(err.Error(), "flag needs an argument: -"); ok {
		color.Red().Printfln("The '%s' option requires a value.", flag)
		return nil
	}
	if errMsg := err.Error(); strings.HasPrefix(errMsg, "invalid value") && strings.Contains(errMsg, "for flag -") {
		var value, flag string
		if _, parseErr := fmt.Sscanf(errMsg, "invalid value %q for flag -%s", &value, &flag); parseErr == nil {
			color.Red().Printfln("Invalid value '%s' for option '%s'.", value, strings.TrimSuffix(flag, ":"))
			return nil
		}
	}

	return err
}

func printHelpCustom(out io.Writer, templ string, data interface{}, _ map[string]interface{}) {
	funcMap := template.FuncMap{
		"capitalize":         capitalize,
		"colorize":           colorize,
		"getFlagName":        getFlagName,
		"getFlagDefaultText": getFlagDefaultText,
		"indent":             indent,
		"join":               strings.Join,
		"offset":             offset,
		"offsetCommands":     offsetCommands,
		"offsetFlags":        offsetFlags,
		"sortFlags":          sortFlags,
		"sortCommands":       sortCommands,
		"subtract":           subtract,
		"trim":               strings.TrimSpace,
		"wrap":               wrap,
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Funcs(colorsFuncMap).Parse(templ))
	templates := map[string]string{
		"usageTemplate":   usageTemplate,
		"commandTemplate": commandTemplate,
		"flagTemplate":    flagTemplate,
	}
	for name, value := range templates {
		if _, err := t.New(name).Parse(value); err != nil {
			if os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != "" {
				_, _ = fmt.Fprintf(cli.ErrWriter, "CLI TEMPLATE ERROR: %#v\n", err)
			}
		}
	}

	if noANSI || env.IsNoANSI() {
		color.Disable()
	}
	err := t.Execute(w, data)
	if err != nil {
		// If the writer is closed, t.Execute will fail, and there's nothing
		// we can do to recover.
		if os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != "" {
			_, _ = fmt.Fprintf(cli.ErrWriter, "CLI TEMPLATE ERROR: %#v\n", err)
		}
		return
	}
	_ = w.Flush()
}

func printVersion(ctx *cli.Context) {
	_, _ = fmt.Fprintf(ctx.App.Writer, "%v %v\n", ctx.App.Usage, color.Green().Sprint(ctx.App.Version))
}

func sortCommands(commands []*cli.Command) []*cli.Command {
	sort.Slice(commands, func(i, j int) bool {
		return strings.Join(commands[i].Names(), ", ") < strings.Join(commands[j].Names(), ", ")
	})
	return commands
}

func sortFlags(flags []cli.Flag) []cli.Flag {
	// built-in flags should be at the end.
	var (
		builtinFlags = map[string]struct{}{
			cli.FlagNamePrefixer(noANSIFlag.Names(), ""):      {},
			cli.FlagNamePrefixer(cli.HelpFlag.Names(), ""):    {},
			cli.FlagNamePrefixer(cli.VersionFlag.Names(), ""): {},
		}
		isBuiltinFlags = func(n string) bool {
			_, ok := builtinFlags[n]
			return ok
		}
	)

	sort.Slice(flags, func(i, j int) bool {
		a, b := cli.FlagNamePrefixer(flags[i].Names(), ""), cli.FlagNamePrefixer(flags[j].Names(), "")
		if isBuiltinFlags(a) && !isBuiltinFlags(b) {
			return false
		}
		if !isBuiltinFlags(a) && isBuiltinFlags(b) {
			return true
		}
		return a < b
	})

	return flags
}

func subtract(a, b int) int {
	return a - b
}

func wrap(input string, offset int) string {
	var ss []string

	lines := strings.Split(input, "\n")

	padding := strings.Repeat(" ", offset)

	for i, line := range lines {
		if line == "" {
			ss = append(ss, line)
		} else {
			wrapped := wrapLine(line, offset, padding)
			if i == 0 {
				ss = append(ss, wrapped)
			} else {
				ss = append(ss, padding+wrapped)

			}
		}
	}

	return strings.Join(ss, "\n")
}

func wrapLine(input string, offset int, padding string) string {
	if maxLineLength <= offset || len(input) <= maxLineLength-offset {
		return input
	}

	lineWidth := maxLineLength - offset
	words := strings.Fields(input)
	if len(words) == 0 {
		return input
	}

	wrapped := words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + padding + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	return wrapped
}
