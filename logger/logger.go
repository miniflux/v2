// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"
	"os"
	"time"
)

// Debug sends a debug log message.
func Debug(format string, v ...interface{}) {
	formatMessage("DEBUG", format, v...)
}

// Info sends an info log message.
func Info(format string, v ...interface{}) {
	formatMessage("INFO", format, v...)
}

// Error sends an error log message.
func Error(format string, v ...interface{}) {
	formatMessage("ERROR", format, v...)
}

// Fatal sends a fatal log message and stop the execution of the program.
func Fatal(format string, v ...interface{}) {
	formatMessage("FATAL", format, v...)
	os.Exit(1)
}

func formatMessage(level, format string, v ...interface{}) {
	prefix := fmt.Sprintf("[%s] [%s] ", time.Now().Format("2006-01-02T15:04:05"), level)
	fmt.Fprintf(os.Stderr, prefix+format+"\n", v...)
}
