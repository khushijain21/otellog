package otellog

import (
	stdlog "log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/logtest"
)

var (
	testMessage = "log message"
	loggerName  = "name"
)

func TestCore(t *testing.T) {
	rec := logtest.NewRecorder()
	logger := stdlog.New(NewOTelWriter(loggerName, WithLoggerProvider(rec)), "logger: ", stdlog.Ltime|stdlog.LUTC)
	logger.Println(testMessage)
	got := rec.Result()[0].Records[0]
	now := time.Now().UTC()
	delta := now.Sub(got.Timestamp()).Seconds()
	// Assert that the timestamp is within an acceptable duration
	assert.InDelta(t, 0, delta, 5, "Logged time is not within 5 seconds of current time")
}

func value2Result(v log.Value) any {
	switch v.Kind() {
	case log.KindBool:
		return v.AsBool()
	case log.KindFloat64:
		return v.AsFloat64()
	case log.KindInt64:
		return v.AsInt64()
	case log.KindString:
		return v.AsString()
	case log.KindBytes:
		return v.AsBytes()
	case log.KindSlice:
		var s []any
		for _, val := range v.AsSlice() {
			s = append(s, value2Result(val))
		}
		return s
	case log.KindMap:
		m := make(map[string]any)
		for _, val := range v.AsMap() {
			m[val.Key] = value2Result(val.Value)
		}
		return m
	}
	return nil
}
