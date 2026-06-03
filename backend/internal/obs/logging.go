// Package obs handles structured logging
package obs

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	service string
}

func NewLogger(service string) *Logger {
	return &Logger{service: service}
}

func (l *Logger) Info(ctx context.Context, msg string, fields map[string]any) {
	l.emit("info", ctx, msg, fields)
}

func (l *Logger) Error(ctx context.Context, msg string, err error, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.emit("error", ctx, msg, fields)
}

func (l *Logger) emit(level string, ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}

	out := map[string]any{
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		"level":   level,
		"msg":     msg,
		"service": l.service,
	}

	// propagate trace_id if present
	if ctx != nil {
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			sc := span.SpanContext()

			out["trace_id"] = sc.TraceID().String()
			out["span_id"] = sc.SpanID().String()
		}
	}

	for k, v := range fields {
		out[k] = v
	}

	_ = json.NewEncoder(os.Stdout).Encode(out)
}
