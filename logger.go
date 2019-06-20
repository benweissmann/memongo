package memongo

import (
	"log"
	"os"
)

type LogLevel int

const (
	LogLevelDebug  = 2
	LogLevelInfo   = 3
	LogLevelWarn   = 4
	LogLevelSilent = 10
)

const defaultLogLevel = LogLevelInfo

type logger struct {
	level LogLevel
	out   *log.Logger
}

func newLogger(out *log.Logger, level LogLevel) *logger {
	if out == nil {
		out = log.New(os.Stdout, "", 0)
	}

	if level == 0 {
		level = defaultLogLevel
	}

	return &logger{
		level: level,
		out:   out,
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if l.level <= LogLevelDebug {
		l.out.Printf("[memongo] [DEBUG] "+format, v...)
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.out.Printf("[memongo] [INFO]  "+format, v...)
	}
}

func (l *logger) Warnf(format string, v ...interface{}) {
	if l.level <= LogLevelWarn {
		l.out.Printf("[memongo] [WARN]  "+format, v...)
	}
}
