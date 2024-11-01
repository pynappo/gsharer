package main

import (
	"log/slog"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
)

func GsharerLogger() *slog.Logger {
	return slog.New(tint.NewHandler(colorable.NewColorableStderr(), nil))
}
