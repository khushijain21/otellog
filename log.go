package otellog

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

type config struct {
	provider  log.LoggerProvider
	version   string
	schemaURL string
}

func newConfig(options []Option) config {
	var c config
	for _, opt := range options {
		c = opt.apply(c)
	}

	if c.provider == nil {
		c.provider = global.GetLoggerProvider()
	}

	return c
}

func (c config) logger(name string) log.Logger {
	var opts []log.LoggerOption
	if c.version != "" {
		opts = append(opts, log.WithInstrumentationVersion(c.version))
	}
	if c.schemaURL != "" {
		opts = append(opts, log.WithSchemaURL(c.schemaURL))
	}
	return c.provider.Logger(name, opts...)
}

// Option configures a [Core].
type Option interface {
	apply(config) config
}

type optFunc func(config) config

func (f optFunc) apply(c config) config { return f(c) }

// WithVersion returns an [Option] that configures the version of the
// [log.Logger] used by a [Core]. The version should be the version of the
// package that is being logged.
func WithVersion(version string) Option {
	return optFunc(func(c config) config {
		c.version = version
		return c
	})
}

// WithSchemaURL returns an [Option] that configures the semantic convention
// schema URL of the [log.Logger] used by a [Core]. The schemaURL should be
// the schema URL for the semantic conventions used in log records.
func WithSchemaURL(schemaURL string) Option {
	return optFunc(func(c config) config {
		c.schemaURL = schemaURL
		return c
	})
}

// WithLoggerProvider returns an [Option] that configures [log.LoggerProvider]
// used by a [Core] to create its [log.Logger].
//
// By default if this Option is not provided, the Handler will use the global
// LoggerProvider.
func WithLoggerProvider(provider log.LoggerProvider) Option {
	return optFunc(func(c config) config {
		c.provider = provider
		return c
	})
}

type OTelWriter struct {
	logger log.Logger
}

func NewOTelWriter(name string, options ...Option) *OTelWriter {
	cfg := newConfig(options)
	return &OTelWriter{logger: cfg.logger(name)}
}

func (o OTelWriter) Write(p []byte) (n int, err error) {
	// Convert the buffer to a string
	logString := string(p)
	rec := log.Record{}

	fmt.Println("logstring", logString)
	// Regex to find UTC timestamp format
	timeRegex := regexp.MustCompile(`([01]\d|2[0-3]):([0-5]\d):([0-5]\d)`)
	matches := timeRegex.FindAllString(logString, -1)

	// This is not going to workk
	// Extract and print timestamps in UTC
	for _, match := range matches {
		layout := "15:04:05"
		t, err := time.Parse(layout, match)
		if err != nil {
			return 0, err
		}
		fmt.Println(t, "T")
		rec.SetTimestamp(t)
	}

	// // Extract other attributes using regular expressions
	// // For example, extract log levels (INFO, ERROR, etc.)
	// levelRegex := regexp.MustCompile(`\b(INFO|ERROR|DEBUG|WARN)\b`)
	// levels := levelRegex.FindAllString(logString, -1)
	// fmt.Println("Log levels found:")
	// for _, level := range levels {
	// 	fmt.Println(level)
	// }

	o.logger.Emit(context.Background(), log.Record{})
	return n, nil
}
