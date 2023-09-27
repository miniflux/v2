// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"io"
	"log/slog"
)

func InitializeDefaultLogger(logLevel string, logFile io.Writer, logFormat string, logTime bool) error {
	var programLogLevel = new(slog.LevelVar)
	switch logLevel {
	case "debug":
		programLogLevel.Set(slog.LevelDebug)
	case "info":
		programLogLevel.Set(slog.LevelInfo)
	case "warning":
		programLogLevel.Set(slog.LevelWarn)
	case "error":
		programLogLevel.Set(slog.LevelError)
	}

	logHandlerOptions := &slog.HandlerOptions{Level: programLogLevel}
	if !logTime {
		logHandlerOptions.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		}
	}

	var logger *slog.Logger
	switch logFormat {
	case "json":
		logger = slog.New(slog.NewJSONHandler(logFile, logHandlerOptions))
	default:
		logger = slog.New(slog.NewTextHandler(logFile, logHandlerOptions))
	}

	slog.SetDefault(logger)

	return nil
}
