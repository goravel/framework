package console

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"

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

func NewList(schedule schedule.Schedule) *List {
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
	events := r.schedule.Events()
	if len(events) == 0 {
		ctx.Warning("No scheduled tasks have been defined.")
		return nil
	}

	cronExpressionSpacing := r.getCronExpressionSpacing(events)

	for _, event := range events {
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
		ctx.TwoColumnDetail(fmt.Sprintf("<fg=yellow>%s</>  %s", r.formatCronExpression(event.GetCron(), cronExpressionSpacing), cmd), nextDue)
	}
	return nil
}

func (r *List) formatCronExpression(expression string, spacing []int) string {
	parts := strings.Fields(expression)
	padded := make([]string, len(spacing))

	for i := 0; i < len(spacing); i++ {
		val := ""
		if i < len(parts) {
			val = parts[i]
		}
		padLen := spacing[i] - utf8.RuneCountInString(val)
		if padLen < 0 {
			padLen = 0
		}
		padded[i] = val + strings.Repeat(" ", padLen)
	}

	return strings.Join(padded, " ")
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

func (r *List) getCronExpressionSpacing(events []schedule.Event) []int {
	var rows [][]int
	for _, event := range events {
		parts := strings.Fields(event.GetCron())
		lengths := make([]int, len(parts))
		for i, part := range parts {
			lengths[i] = utf8.RuneCountInString(part)
		}
		rows = append(rows, lengths)
	}
	if len(rows) == 0 {
		return []int{}
	}

	spacing := make([]int, len(rows[0]))
	for _, row := range rows {
		for i, length := range row {
			if length > spacing[i] {
				spacing[i] = length
			}
		}
	}

	return spacing
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
