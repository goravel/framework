package telemetry

import (
	"strings"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
)

const (
	propagatorTraceContext = "tracecontext"
	propagatorBaggage      = "baggage"
	propagatorB3           = "b3"
	propagatorB3Multi      = "b3multi"
	propagatorNone         = "none"
)

var (
	defaultCompositePropagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	nonePropagator = propagation.NewCompositeTextMapPropagator()
)

func newCompositeTextMapPropagator(nameStr string) propagation.TextMapPropagator {
	if nameStr == "" {
		return defaultCompositePropagator
	}

	if nameStr == propagatorNone {
		return nonePropagator
	}

	var propagators []propagation.TextMapPropagator
	for _, name := range strings.Split(nameStr, ",") {
		switch strings.TrimSpace(name) {
		case propagatorTraceContext:
			propagators = append(propagators, propagation.TraceContext{})
		case propagatorBaggage:
			propagators = append(propagators, propagation.Baggage{})
		case propagatorB3:
			propagators = append(propagators, b3.New(b3.WithInjectEncoding(b3.B3SingleHeader)))
		case propagatorB3Multi:
			propagators = append(propagators, b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)))
		}
	}

	if len(propagators) == 0 {
		return defaultCompositePropagator
	}

	return propagation.NewCompositeTextMapPropagator(propagators...)
}
