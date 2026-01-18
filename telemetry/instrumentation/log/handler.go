package log

import (
	"context"
	"sync"
	"time"

	otellog "go.opentelemetry.io/otel/log"

	contractslog "github.com/goravel/framework/contracts/log"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

var _ contractslog.Handler = (*handler)(nil)

type handler struct {
	resolver       contractstelemetry.Resolver  // The un-executed function
	telemetry      contractstelemetry.Telemetry // The cached instance
	enabled        bool
	instrumentName string
	logger         otellog.Logger
	mu             sync.Mutex
}

func (r *handler) Enabled(level contractslog.Level) bool {
	return r.enabled
}

func (r *handler) Handle(entry contractslog.Entry) error {
	if !r.enabled {
		return nil
	}

	logger := r.getLogger()
	if logger == nil {
		return nil
	}

	ctx := entry.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	logger.Emit(ctx, r.convertEntry(entry))

	return nil
}

func (r *handler) getLogger() otellog.Logger {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.logger != nil {
		return r.logger
	}

	if r.telemetry == nil && r.resolver != nil {
		r.telemetry = r.resolver()
	}

	if r.telemetry != nil {
		r.logger = r.telemetry.Logger(r.instrumentName)
	}

	return r.logger
}

func (r *handler) convertEntry(e contractslog.Entry) otellog.Record {
	var record otellog.Record
	record.SetTimestamp(e.Time())
	record.SetObservedTimestamp(time.Now())
	record.SetBody(otellog.StringValue(e.Message()))
	record.SetSeverity(toSeverity(e.Level()))
	record.SetSeverityText(e.Level().String())

	estimatedSize := 5 + len(e.With()) + len(e.Data())
	attrs := make([]otellog.KeyValue, 0, estimatedSize)

	if code := e.Code(); code != "" {
		attrs = append(attrs, otellog.String("code", code))
	}
	if domain := e.Domain(); domain != "" {
		attrs = append(attrs, otellog.String("domain", domain))
	}
	if hint := e.Hint(); hint != "" {
		attrs = append(attrs, otellog.String("hint", hint))
	}
	if owner := e.Owner(); owner != nil {
		attrs = append(attrs, otellog.KeyValue{Key: "owner", Value: toValue(owner)})
	}
	if user := e.User(); user != nil {
		attrs = append(attrs, otellog.KeyValue{Key: "user", Value: toValue(user)})
	}
	if tags := e.Tags(); len(tags) > 0 {
		attrs = append(attrs, otellog.KeyValue{Key: "tags", Value: toValue(tags)})
	}
	if req := e.Request(); len(req) > 0 {
		attrs = append(attrs, otellog.KeyValue{Key: "request", Value: toValue(req)})
	}
	if res := e.Response(); len(res) > 0 {
		attrs = append(attrs, otellog.KeyValue{Key: "response", Value: toValue(res)})
	}
	if tr := e.Trace(); len(tr) > 0 {
		attrs = append(attrs, otellog.KeyValue{Key: "trace", Value: toValue(tr)})
	}

	for k, v := range e.With() {
		attrs = append(attrs, otellog.KeyValue{Key: k, Value: toValue(v)})
	}

	for k, v := range e.Data() {
		// Goravel packs all structured metadata (trace, request, user, etc.) into the "root" key.
		// Since we have already extracted and mapped these fields to top-level OTel attributes above,
		// we skip "root" here to prevent duplicating the entire context map.
		if k != "root" {
			attrs = append(attrs, otellog.KeyValue{Key: k, Value: toValue(v)})
		}
	}

	record.AddAttributes(attrs...)
	return record
}
