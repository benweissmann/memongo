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
