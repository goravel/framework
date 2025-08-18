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
)

func init() {
	cli.HelpPrinterCustom = printHelpCustom
	cli.AppHelpTemplate = appHelpTemplate
	cli.CommandHelpTemplate = commandHelpTemplate
	cli.VersionPrinter = printVersion
	huh.ErrUserAborted = cli.Exit(color.Red().Sprint("Cancelled."), 0)
}

const maxLineLength = 10000

var usageTemplate = `{{if .UsageText}}{{wrap (colorize .UsageText) 3}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}}{{if .ArgsUsage}}{{.ArgsUsage}}{{else}}{{if .Args}} [arguments...]{{end}}{{end}}{{end}}`
var commandTemplate = `{{ $cv := offsetCommands .VisibleCommands 5}}{{range .VisibleCategories}}{{if .Name}}
 {{yellow .Name}}:{{end}}{{range .VisibleCommands}}
  {{$s := join .Names ", "}}{{green $s}}{{ $sp := subtract $cv (offset $s 3) }}{{ indent $sp ""}}{{wrap (colorize .Usage) $cv}}{{end}}{{end}}`

var flagTemplate = `{{ $cv := offsetFlags .VisibleFlags 5}}{{range  .VisibleFlags}}
   {{$s := getFlagName .}}{{green $s}}{{ $sp := subtract $cv (offset $s 1) }}{{ indent $sp ""}}{{$us := (capitalize .Usage)}}{{wrap (colorize $us) $cv}}{{$df := getFlagDefaultText . }}{{if $df}} {{yellow $df}}{{end}}{{end}}`

var appHelpTemplate = `{{$v := offset .Usage 6}}{{wrap (colorize .Usage) 3}}{{if .Version}} {{green (wrap .Version $v)}}{{end}}

{{ yellow "Usage:" }}
   {{if .UsageText}}{{wrap (colorize .UsageText) 3}}{{end}}{{if .VisibleFlags}}

{{ yellow "Options:" }}{{template "flagTemplate" .}}{{end}}{{if .VisibleCommands}}

{{ yellow "Available commands:" }}{{template "commandTemplate" .}}{{end}}
`

var commandHelpTemplate = `{{ yellow "Description:" }}
   {{ (colorize .Usage) }}

{{ yellow "Usage:" }}
   {{template "usageTemplate" .}}{{if .VisibleFlags}}

{{ yellow "Options:" }}{{template "flagTemplate" .}}{{end}}
`

var colorsFuncMap = template.FuncMap{
	"green":   color.Green().Sprint,
	"red":     color.Red().Sprint,
	"blue":    color.Blue().Sprint,
	"yellow":  color.Yellow().Sprint,
	"cyan":    color.Cyan().Sprint,
	"white":   color.White().Sprint,
	"gray":    color.Gray().Sprint,
	"default": color.Default().Sprint,
	"black":   color.Black().Sprint,
	"magenta": color.Magenta().Sprint,
}

func subtract(a, b int) int {
	return a - b
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.ReplaceAll(v, "\n", "\n"+pad)
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

func getFlagName(flag cli.DocGenerationFlag) string {
	names := flag.Names()
	sort.Slice(names, func(i, j int) bool {
		return len(names[i]) < len(names[j])
	})

	return cli.FlagNamePrefixer(names, "")
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

func printVersion(ctx *cli.Context) {
	_, _ = fmt.Fprintf(ctx.App.Writer, "%v %v\n", ctx.App.Usage, color.Green().Sprint(ctx.App.Version))
}

func printHelpCustom(out io.Writer, templ string, data interface{}, _ map[string]interface{}) {

	funcMap := template.FuncMap{
		"join":               strings.Join,
		"subtract":           subtract,
		"indent":             indent,
		"trim":               strings.TrimSpace,
		"capitalize":         capitalize,
		"wrap":               wrap,
		"offset":             offset,
		"offsetCommands":     offsetCommands,
		"offsetFlags":        offsetFlags,
		"getFlagName":        getFlagName,
		"getFlagDefaultText": getFlagDefaultText,
		"colorize":           colorize,
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
