package grq

import (
	"runtime"

	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

func attachCodeLocationToSpan(span trace.Span) {
	_, file, line, ok := runtime.Caller(1) // verify
	if ok {
		span.SetAttributes(semconv.CodeFilePath(file), semconv.CodeLineNumber(line))
	}
}
