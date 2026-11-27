package console

import (
	"cmp"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
)

type AboutCommand struct {
	app foundation.Application
}

type information struct {
	section map[string]int
	details [][]foundation.AboutItem
}

// aboutMu guards appInformation and customInformationResolvers. It guards each
// individual read or write. Callbacks and resolvers always run outside it.
var aboutMu sync.RWMutex
var appInformation = &information{section: make(map[string]int)}
var customInformationResolvers []func()
var getGoVersion = runtime.Version

func NewAboutCommand(app foundation.Application) *AboutCommand {
	return &AboutCommand{
		app: app,
	}
}

// Signature The name and signature of the console command.
func (r *AboutCommand) Signature() string {
	return "about"
}

// Description The console command description.
func (r *AboutCommand) Description() string {
	return "Display basic information about your application"
}

// Extend The console command extend.
func (r *AboutCommand) Extend() command.Extend {
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
func (r *AboutCommand) Handle(ctx console.Context) error {
	r.gatherApplicationInformation()
	ctx.NewLine()
	appInformation.Range(ctx.Option("only"), func(section string, items []foundation.AboutItem) {
		ctx.TwoColumnDetail("<fg=green;op=bold>"+section+"</>", "")
		for i := range items {
			ctx.TwoColumnDetail(items[i].Key, items[i].Value)
		}
		ctx.NewLine()
	})
	return nil
}

// gatherApplicationInformation Gather information about the application.
func (r *AboutCommand) gatherApplicationInformation() {
	configFacade := r.app.MakeConfig()
	appInformation.addToSection("Environment", []foundation.AboutItem{
		{Key: "Application Name", Value: configFacade.GetString("app.name")},
		{Key: "Goravel Version", Value: strings.TrimPrefix(r.app.Version(), "v")},
		{Key: "Go Version", Value: strings.TrimPrefix(getGoVersion(), "go")},
		{Key: "Environment", Value: configFacade.GetString("app.env")},
		{Key: "Debug Mode", Value: func() string {
			mode := "OFF"
			if configFacade.GetBool("app.debug") {
				mode = "<fg=yellow;op=bold>ENABLED</>"
			}
			return mode
		}()},
		{Key: "URL", Value: configFacade.GetString("http.url")},
		{Key: "HTTP Host", Value: configFacade.GetString("http.host")},
		{Key: "HTTP Port", Value: configFacade.GetString("http.port")},
		{Key: "GRPC Host", Value: configFacade.GetString("grpc.host")},
		{Key: "GRPC Port", Value: configFacade.GetString("grpc.port")},
	})
	appInformation.addToSection("Drivers", []foundation.AboutItem{
		{Key: "Cache", Value: configFacade.GetString("cache.default")},
		{Key: "Database", Value: configFacade.GetString("database.default")},
		{Key: "Hashing", Value: configFacade.GetString("hashing.driver")},
		{Key: "Http", Value: configFacade.GetString("http.default")},
		{Key: "Logs", Value: func() string {
			logs := configFacade.GetString("logging.default")
			if logChannel := logs; configFacade.GetString("logging.channels."+logChannel+".driver") == "stack" {
				if secondary, ok := configFacade.Get("logging.channels." + logChannel + ".channels").([]string); ok {
					logs = fmt.Sprintf("<fg=yellow;op=bold>%s</> <fg=gray;op=bold>/</> %s", logChannel, strings.Join(secondary, ", "))
				}
			}
			return logs
		}()},
		{Key: "Mail", Value: configFacade.GetString("mail.default", "smtp")},
		{Key: "Queue", Value: configFacade.GetString("queue.default")},
		{Key: "Session", Value: configFacade.GetString("session.default")},
	})
	// Copy the resolvers under the lock and run them outside it, because they write to the same state.
	aboutMu.RLock()
	resolvers := slices.Clone(customInformationResolvers)
	aboutMu.RUnlock()
	for i := range resolvers {
		resolvers[i]()
	}
}

// addToSection Add a new section to the application information.
func (info *information) addToSection(section string, items []foundation.AboutItem) {
	aboutMu.Lock()
	defer aboutMu.Unlock()

	index, ok := info.section[section]
	if !ok {
		index = len(info.details)
		info.section[section] = index
		info.details = append(info.details, make([]foundation.AboutItem, 0))
	}
	info.details[index] = append(info.details[index], items...)
}

// Range Iterate over the application information sections.
func (info *information) Range(section string, ranger func(s string, items []foundation.AboutItem)) {
	type sectionItems struct {
		name  string
		index int
		items []foundation.AboutItem
	}

	// Copy the sections and their items under the lock and call the callback outside it.
	var sections []sectionItems
	aboutMu.RLock()
	for s, index := range info.section {
		if len(section) == 0 || strings.EqualFold(section, s) {
			sections = append(sections, sectionItems{name: s, index: index, items: slices.Clone(info.details[index])})
		}
	}
	aboutMu.RUnlock()

	if len(sections) > 1 {
		slices.SortFunc(sections, func(a, b sectionItems) int {
			return cmp.Compare(a.index, b.index)
		})
	}
	for i := range sections {
		ranger(sections[i].name, sections[i].items)
	}
}

// AddAboutInformation Add custom information to the application information.
//
// Add entries from a service provider Register or Boot, or from an application builder callback,
// because those run again when the application is rebuilt. Refresh flushes the application
// information and the registered entries, so an entry added once outside the build is lost on refresh.
func AddAboutInformation(section string, items ...foundation.AboutItem) {
	aboutMu.Lock()
	defer aboutMu.Unlock()

	customInformationResolvers = append(customInformationResolvers, func() {
		appInformation.addToSection(section, items)
	})
}

// FlushAboutState Clear the collected application information and the custom information resolvers.
// It is safe on empty state and can run repeatedly.
func FlushAboutState() {
	aboutMu.Lock()
	defer aboutMu.Unlock()

	// Clear in place, so that readers of appInformation never see the variable change.
	clear(appInformation.section)
	appInformation.details = nil
	customInformationResolvers = nil
}
