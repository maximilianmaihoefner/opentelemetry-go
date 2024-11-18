package otelzerolog

/*
Copyright Agoda Services Co.,Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/rs/zerolog"
	logs "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdk "go.opentelemetry.io/otel/sdk/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	instrumentationName = "github.com/agoda-com/opentelemetry-go/otelzerolog"
)

var instrumentationScope = instrumentation.Scope{
	Name:      instrumentationName,
	Version:   Version(),
	SchemaURL: semconv.SchemaURL,
}

type Hook struct {
	logs.Logger
}

var _ zerolog.Hook = (*Hook)(nil)

func NewHook(loggerProvider *sdk.LoggerProvider) *Hook {
	logger := loggerProvider.Logger(
		instrumentationScope.Name,
		logs.WithInstrumentationVersion(instrumentationScope.Version),
	)
	return &Hook{logger}
}

func (h Hook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if !e.Enabled() {
		return
	}

	ctx := e.GetCtx()
	//span := trace.SpanFromContext(ctx).SpanContext()

	//var spanID trace.SpanID
	//var traceID trace.TraceID
	//var traceFlags trace.TraceFlags
	//if span.IsValid() {
	//	spanID = span.SpanID()
	//	traceID = span.TraceID()
	//	traceFlags = span.TraceFlags()
	//}

	now := time.Now()
	//severityText := otelLevelText(level)
	severityNumber := otelLevelNumber(level)

	logData := make(map[string]interface{})
	// create a string that appends } to the end of the buf variable you access via reflection
	ev := fmt.Sprintf("%s}", reflect.ValueOf(e).Elem().FieldByName("buf"))
	_ = json.Unmarshal([]byte(ev), &logData)

	// TODO: this is very hacky, but it works for now
	var attributes []logs.KeyValue
	for k, v := range logData {
		attributes = append(attributes, otelAttribute(k, v)...)
	}

	//lrc := otel.LogRecordConfig{
	//	Timestamp:            &now,
	//	ObservedTimestamp:    now,
	//	TraceId:              &traceID,
	//	SpanId:               &spanID,
	//	TraceFlags:           &traceFlags,
	//	SeverityText:         &severityText,
	//	SeverityNumber:       &severityNumber,
	//	BodyAny:              &msg,
	//	Resource:             nil,
	//	InstrumentationScope: &instrumentationScope,
	//	Attributes:           &attributes,
	//}
	//r := otel.NewLogRecord(lrc)

	r := logs.Record{}
	r.SetTimestamp(now)
	r.SetObservedTimestamp(now)
	r.SetSeverity(logs.Severity(severityNumber))
	r.SetBody(logs.StringValue(msg))
	r.AddAttributes(attributes...)
	h.Emit(ctx, r)
}
