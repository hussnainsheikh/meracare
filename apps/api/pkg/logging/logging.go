// Package logging configures the structured logger used across the API.
//
// Care and health information is sensitive (docs/09-security-privacy.md).
// Never log access tokens, passwords, or health data. Log identifiers and
// outcomes, not payloads.
package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// contextKey is unexported so no other package can collide with it.
type contextKey struct{}

var loggerKey contextKey

// Options configures the root logger.
type Options struct {
	// Level is one of debug, info, warn, error. Defaults to info.
	Level string
	// Development switches to human-readable text output.
	Development bool
	// ServiceName is attached to every record.
	ServiceName string
}

// New builds the root logger.
func New(w io.Writer, opts Options) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{Level: ParseLevel(opts.Level)}

	var handler slog.Handler
	if opts.Development {
		handler = slog.NewTextHandler(w, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(w, handlerOpts)
	}

	logger := slog.New(handler)
	if opts.ServiceName != "" {
		logger = logger.With(slog.String("service", opts.ServiceName))
	}
	return logger
}

// ParseLevel maps a configuration string to a slog level, defaulting to info.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithLogger returns a context carrying the given logger.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the request-scoped logger, or the default logger when the
// context does not carry one.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}
