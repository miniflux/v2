// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package logger // import "miniflux.app/logger"

import (
	"fmt"
	"os"
	"time"
)

var requestedLevel = InfoLevel
var displayDateTime = false

// LogLevel type.
type LogLevel uint32

const (
	// FatalLevel should be used in fatal situations, the app will exit.
	FatalLevel LogLevel = iota

	// ErrorLevel should be used when someone should really look at the error.
	ErrorLevel

	// InfoLevel should be used during normal operations.
	InfoLevel

	// DebugLevel should be used only during development.
	DebugLevel
)

func (level LogLevel) String() string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// EnableDateTime enables date time in log messages.
func EnableDateTime() {
	displayDateTime = true
}

// EnableDebug increases logging, more verbose (debug)
func EnableDebug() {
	requestedLevel = DebugLevel
	formatMessage(InfoLevel, "Debug mode enabled")
}

// Debug sends a debug log message.
func Debug(format string, v ...interface{}) {
	if requestedLevel >= DebugLevel {
		formatMessage(DebugLevel, format, v...)
	}
}

// Info sends an info log message.
func Info(format string, v ...interface{}) {
	if requestedLevel >= InfoLevel {
		formatMessage(InfoLevel, format, v...)
	}
}

// Error sends an error log message.
func Error(format string, v ...interface{}) {
	if requestedLevel >= ErrorLevel {
		formatMessage(ErrorLevel, format, v...)
	}
}

// Fatal sends a fatal log message and stop the execution of the program.
func Fatal(format string, v ...interface{}) {
	if requestedLevel >= FatalLevel {
		formatMessage(FatalLevel, format, v...)
		os.Exit(1)
	}
}

func formatMessage(level LogLevel, format string, v ...interface{}) {
	var prefix string

	if displayDateTime {
		prefix = fmt.Sprintf("[%s] [%s] ", time.Now().Format("2006-01-02T15:04:05"), level)
	} else {
		prefix = fmt.Sprintf("[%s] ", level)
	}

	fmt.Fprintf(os.Stderr, prefix+format+"\n", v...)
}
