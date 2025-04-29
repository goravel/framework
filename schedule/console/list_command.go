package console

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"

	"github.com/robfig/cron/v3"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

var cronParser = cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

type List struct {
	schedule schedule.Schedule
}

func NewListCommand(schedule schedule.Schedule) *List {
	return &List{
		schedule: schedule,
	}
}

// Signature The name and signature of the console command.
func (r *List) Signature() string {
	return "schedule:list"
}

// Description The console command description.
func (r *List) Description() string {
	return "List all scheduled tasks"
}

// Extend The console command extend.
func (r *List) Extend() command.Extend {
	return command.Extend{
		Category: "schedule",
	}
}

// Handle Execute the console command.
func (r *List) Handle(ctx console.Context) error {
	ctx.NewLine()
	for _, event := range r.schedule.Events() {
		cmd := event.GetName()

		// display artisan command signature,when it has command
		if c := event.GetCommand(); c != "" {
			cmd = "artisan " + c
			// highlight the parameters...
			cmd = regexp.MustCompile(`(artisan [\w\-:]+) (.+)`).ReplaceAllString(cmd, `$1 <fg=yellow>$2</>`)
		}

		// display closure location,when it doesn't have a name
		if len(cmd) == 0 && event.GetCallback() != nil {
			file, line := r.getClosureLocation(event.GetCallback())
			file, _ = filepath.Rel(str.Of(file).Dirname(3).String(), file)
			cmd = fmt.Sprintf("Closure at: %s:%d", file, line)
		}

		var nextDue string
		if nd := r.getNextDueDate(event.GetCron()); len(nd) > 0 {
			nextDue = fmt.Sprintf("<fg=7472a3>Next Due: %s</>", nd)
		}
		ctx.TwoColumnDetail(fmt.Sprintf("<fg=yellow>%s</>  %s", event.GetCron(), cmd), nextDue)
	}
	return nil
}

func (r *List) getClosureLocation(closure any) (file string, line int) {
	v := reflect.ValueOf(closure)
	if v.Kind() == reflect.Func {
		ptr := v.Pointer()
		if fn := runtime.FuncForPC(ptr); fn != nil {
			file, line = fn.FileLine(ptr)
		}
	}

	return
}

func (r *List) getNextDueDate(cronSpec string) string {
	if sch, err := cronParser.Parse(cronSpec); err == nil {
		now := carbon.Now()
		if next := sch.Next(now.StdTime()); !next.IsZero() {
			return carbon.FromStdTime(next).DiffForHumans(now)
		}
	}

	return ""
}
