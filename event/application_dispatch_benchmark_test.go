package event

import (
	"strconv"
	"testing"

	"github.com/goravel/framework/contracts/event"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func benchmarkApplication(b *testing.B, patterns int) *Application {
	app := NewApplication(mocksqueue.NewQueue(b))

	noop := func(evt any, args ...any) error { return nil }
	if err := app.Listen("user.created", noop); err != nil {
		b.Fatal(err)
	}
	for i := range patterns {
		if err := app.Listen("segment"+strconv.Itoa(i)+".*", noop); err != nil {
			b.Fatal(err)
		}
	}

	return app
}

// The event names of a real application are not all known up front, so the
// wildcard patterns are benchmarked against both a repeated name and names of
// high cardinality.
func BenchmarkDispatch(b *testing.B) {
	args := []event.Arg{{Type: "string", Value: "goravel"}}

	for _, patterns := range []int{0, 1, 8, 32, 128} {
		name := strconv.Itoa(patterns) + "Patterns"

		b.Run(name+"/RepeatedName", func(b *testing.B) {
			app := benchmarkApplication(b, patterns)
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				app.Dispatch("user.created", args)
			}
		})

		b.Run(name+"/UniqueNames", func(b *testing.B) {
			app := benchmarkApplication(b, patterns)
			b.ReportAllocs()
			b.ResetTimer()

			i := 0
			for b.Loop() {
				app.Dispatch("user."+strconv.Itoa(i), args)
				i++
			}
		})
	}
}
