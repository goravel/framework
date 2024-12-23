package console

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/str"
)

type AboutCommand struct {
	app foundation.Application
}

type information struct {
	section map[string]int
	details [][]kv
}
type kv struct {
	key   string
	value string
}

var appInformation = &information{section: make(map[string]int)}
var customInformationResolvers []func()

func NewAboutCommand(app foundation.Application) *AboutCommand {
	return &AboutCommand{
		app: app,
	}
}

// Signature The name and signature of the console command.
func (receiver *AboutCommand) Signature() string {
	return "about"
}

// Description The console command description.
func (receiver *AboutCommand) Description() string {
	return "Display basic information about your application"
}

// Extend The console command extend.
func (receiver *AboutCommand) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "only",
				Usage: "The section to display",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *AboutCommand) Handle(ctx console.Context) error {
	receiver.gatherApplicationInformation()
	ctx.NewLine()
	appInformation.Range(ctx.Option("only"), func(section string, details []kv) {
		ctx.TwoColumnDetail("<fg=green;op=bold>"+section+"</>", "")
		for i := range details {
			ctx.TwoColumnDetail(details[i].key, details[i].value)
		}
		ctx.NewLine()
	})
	return nil
}

// gatherApplicationInformation Gather information about the application.
func (receiver *AboutCommand) gatherApplicationInformation() {
	configFacade := receiver.app.MakeConfig()
	appInformation.addToSection("Environment",
		"Application Name", configFacade.GetString("app.name"),
		"Goravel Version", str.Of(receiver.app.Version()).LTrim("v").String(),
		"Go Version", str.Of(runtime.Version()).LTrim("go").String(),
		"Environment", configFacade.GetString("app.env"),
		"Debug Mode", func() string {
			if configFacade.GetBool("app.debug") {
				return "<fg=yellow;op=bold>ENABLED</>"
			}
			return "OFF"
		}(),
		"URL", str.Of(configFacade.GetString("http.url")).Replace("http://", "").Replace("https://", "").String(),
		"HTTP Host", configFacade.GetString("http.host"),
		"HTTP Port", configFacade.GetString("http.port"),
	)
	appInformation.addToSection("Drivers",
		"Cache", configFacade.GetString("cache.default"),
		"Database", configFacade.GetString("database.default"),
		"Hashing", configFacade.GetString("hashing.driver"),
		"Http", configFacade.GetString("http.default"),
		"Logs", func() string {
			logChannel := configFacade.GetString("logging.default")
			if configFacade.GetString("logging.channels."+logChannel+".driver") == "stack" {
				if secondary, ok := configFacade.Get("logging.channels." + logChannel + ".channels").([]string); ok {
					return fmt.Sprintf("<fg=yellow;op=bold>%s</> <fg=gray;op=bold>/</> %s", logChannel, strings.Join(secondary, ", "))
				}
			}
			return logChannel
		}(),
		"Mail", configFacade.GetString("mail.default", "smtp"),
		"Queue", configFacade.GetString("queue.default"),
		"Session", configFacade.GetString("session.driver"),
	)
	for i := range customInformationResolvers {
		customInformationResolvers[i]()
	}
}

// addToSection Add a new section to the application information.
func (info *information) addToSection(section, key, vale string, more ...string) {
	index, ok := info.section[section]
	if !ok {
		index = len(info.details)
		info.section[section] = index
		info.details = append(info.details, make([]kv, 0))
	}
	info.details[index] = append(info.details[index], kv{key, vale})
	for i := 0; i < len(more); i += 2 {
		detail := kv{key: more[i]}
		if i+1 < len(more) {
			detail.value = more[i+1]
		}
		info.details[index] = append(info.details[index], detail)
	}
}

// Range Iterate over the application information sections.
func (info *information) Range(section string, ranger func(s string, details []kv)) {
	var sections []string
	for s := range info.section {
		if len(section) == 0 || strings.EqualFold(section, s) {
			sections = append(sections, s)
		}
	}
	if len(sections) > 1 {
		sort.Slice(sections, func(i, j int) bool {
			return info.section[sections[i]] < info.section[sections[j]]
		})
	}
	for i := range sections {
		ranger(sections[i], info.details[info.section[sections[i]]])
	}

}

// AddAboutInformation Add custom information to the application information.
func AddAboutInformation(section, key, value string, more ...string) {
	customInformationResolvers = append(customInformationResolvers, func() {
		appInformation.addToSection(section, key, value, more...)
	})
}
