package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type GsharerLogHandler struct{}

func (h GsharerLogHandler) Enabled(context context.Context, level slog.Level) bool {
	switch level {
	case slog.LevelDebug:
		fallthrough
	case slog.LevelInfo:
		fallthrough
	case slog.LevelWarn:
		fallthrough
	case slog.LevelError:
		return true
	default:
		panic("unreachable")
	}
}

func (h GsharerLogHandler) Handle(context context.Context, record slog.Record) error {
	message := record.Message

	record.Attrs(func(attr slog.Attr) bool {
		message += fmt.Sprintf(" %v", attr)
		return true
	})

	switch record.Level {
	case slog.LevelDebug:
		fallthrough
	case slog.LevelInfo:
		fallthrough
	case slog.LevelWarn:
		fallthrough
	case slog.LevelError:
		fmt.Fprintf(os.Stderr, "[%v]", record.Level, message)
	default:
		panic("unreachable")
	}

	return nil
}

// for advanced users
func (h GsharerLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("unimplemented")
}

// for advanced users
func (h GsharerLogHandler) WithGroup(name string) slog.Handler {
	panic("unimplemented")
}
