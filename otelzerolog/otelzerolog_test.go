package otelzerolog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	sdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type User struct{}

func (u *User) String() string {
	return "I am a user"
}

// configure common attributes for all logs
func newResource() *resource.Resource {
	hostName, _ := os.Hostname()
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("otelzerolog-example"),
		semconv.ServiceVersion("1.0.0"),
		semconv.HostName(hostName),
	)
}

func setupLogProvider() (*bytes.Buffer, *sdk.LoggerProvider) {
	var output bytes.Buffer

	logExporter, _ := stdoutlog.New(stdoutlog.WithWriter(&output))
	loggerProvider := sdk.NewLoggerProvider(
		// use syncer to make sure all logs are flushed before test ends
		sdk.WithProcessor(sdk.NewSimpleProcessor(logExporter)),
		sdk.WithResource(newResource()),
	)

	return &output, loggerProvider
}

func TestZerologHook(t *testing.T) {
	ctx := context.Background()
	output, loggerProvider := setupLogProvider()

	hook := NewHook(loggerProvider)
	logger := log.Hook(hook)
	logger.Info().Ctx(ctx).Str("key", "value").Msg("hello zerolog")

	_ = loggerProvider.Shutdown(ctx)

	actual := output.String()
	assert.Contains(t, actual, "INFO")                                                               // ensure th logger level
	assert.Contains(t, actual, "hello zerolog")                                                      // ensure the message
	assert.Contains(t, actual, "scopeInfo: github.com/agoda-com/opentelemetry-go/otelzerolog:0.0.1") // ensure the scope info
	assert.Contains(t, actual, "service.name=otelzerolog-example")                                   // ensure the resource attributes
	assert.Contains(t, actual, "service.version=1.0.0")                                              // ensure the resource attributes
	assert.Contains(t, actual, "level=info")                                                         // ensure the severity attributes
	assert.Contains(t, actual, "key=value")                                                          // ensure the logger fields
}

func TestZerologHook_ValidSpan(t *testing.T) {
	output, loggerProvider := setupLogProvider()

	// create a span
	tracerProvider := oteltrace.NewTracerProvider(oteltrace.WithResource(newResource()))
	tracer := tracerProvider.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test")
	defer span.End()

	hook := NewHook(loggerProvider)
	logger := log.Hook(hook)
	logger.Info().Ctx(ctx).
		Str("key", "value").
		Strs("strs", []string{"1", "2", "3"}).
		Stringer("stringer", &User{}).
		Int("int", 0).
		Int16("i16", 16).
		Int32("i32", 32).
		Int64("i64", 64).
		Dur("dur", 1).
		Uint("u", 0).
		Uint8("u", 0).
		Uint16("u", 0).
		Uint32("u", 0).
		Uint64("u", 0).
		Float32("float32", 32.32).
		Float64("float64", 64.64).
		Bool("bool", true).
		Interface("interface", &User{}).
		Interface("array", []interface{}{"1", 1, "2", 2, "3", 3}).
		Err(fmt.Errorf("new error")).
		Msg("hello zerolog")

	actual := output.String()
	assert.Contains(t, actual, span.SpanContext().SpanID().String())  // ensure the spanID is logged
	assert.Contains(t, actual, span.SpanContext().TraceID().String()) // ensure the traceID is logged

	logger.Error().Ctx(ctx).Str("key", "value").Discard().Msg("this should not be logged")
	_ = loggerProvider.Shutdown(ctx)
}
