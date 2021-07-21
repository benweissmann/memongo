// Copyright 2021 Tryvium Travels LTD
// Copyright 2019-2020 Ben Weissmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memongolog

import (
	"log"
	"os"
)

// LogLevel is a logging vebosity level
type LogLevel int

const (
	// LogLevelDebug logs all messages
	LogLevelDebug = 2

	// LogLevelInfo logs a small number of information messages
	LogLevelInfo = 3

	// LogLevelWarn only logs messages that indicate a potential problem
	LogLevelWarn = 4

	// LogLevelSilent logs no messages
	LogLevelSilent = 10
)

const defaultLogLevel = LogLevelInfo

// Logger is a logger that filters by log level
type Logger struct {
	level LogLevel
	out   *log.Logger
}

// New constructs a new logger
func New(out *log.Logger, level LogLevel) *Logger {
	if out == nil {
		out = log.New(os.Stdout, "", 0)
	}

	if level == 0 {
		level = defaultLogLevel
	}

	return &Logger{
		level: level,
		out:   out,
	}
}

// Debugf logs at the debug level
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= LogLevelDebug {
		l.out.Printf("[memongo] [DEBUG] "+format, v...)
	}
}

// Infof logs at the info level
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.out.Printf("[memongo] [INFO]  "+format, v...)
	}
}

// Warnf logs at the warning level
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= LogLevelWarn {
		l.out.Printf("[memongo] [WARN]  "+format, v...)
	}
}
