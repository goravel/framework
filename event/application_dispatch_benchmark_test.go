package event

import (
	"strconv"
	"testing"

	"github.com/goravel/framework/contracts/event"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

// benchmarkApplication registers an exact listener plus the given number of
// wildcard patterns, half of which match the dispatched names, so that the cost
// of matching and of collecting the matched listeners is both measured.
func benchmarkApplication(b *testing.B, patterns int) *Application {
	app := NewApplication(mocksqueue.NewQueue(b))

	noop := func(evt any, args ...any) error { return nil }
	if err := app.Listen("user.created", noop); err != nil {
		b.Fatal(err)
	}

	for i := range patterns {
		pattern := "user.*"
		if i%2 == 1 {
			pattern = "segment" + strconv.Itoa(i) + ".*"
		}

		if err := app.Listen(pattern, noop); err != nil {
			b.Fatal(err)
		}
	}

	return app
}

// benchmarkNames builds the dispatched names up front, a name built inside the
// loop would be measured as if it were the cost of dispatching.
func benchmarkNames(count int) []string {
	names := make([]string, count)
	for i := range names {
		names[i] = "user." + strconv.Itoa(i)
	}

	return names
}

func BenchmarkDispatch(b *testing.B) {
	args := []event.Arg{{Type: "string", Value: "goravel"}}
	names := benchmarkNames(1024)

	for _, patterns := range []int{0, 1, 8, 32, 128} {
		name := strconv.Itoa(patterns) + "Patterns"

		b.Run(name+"/RepeatedName", func(b *testing.B) {
			app := benchmarkApplication(b, patterns)
			b.ReportAllocs()

			for b.Loop() {
				app.Dispatch("user.created", args)
			}
		})

		b.Run(name+"/UniqueNames", func(b *testing.B) {
			app := benchmarkApplication(b, patterns)
			b.ReportAllocs()

			i := 0
			for b.Loop() {
				app.Dispatch(names[i%len(names)], args)
				i++
			}
		})
	}
}

func BenchmarkMatchWildcard(b *testing.B) {
	b.Run("Hit", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			matchWildcard("user.*", "user.created")
		}
	})

	b.Run("Miss", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			matchWildcard("segment.*", "user.created")
		}
	})
}
