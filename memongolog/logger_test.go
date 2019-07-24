package memongolog

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	tests := map[string]struct {
		loggerLevel  LogLevel
		messageLevel LogLevel

		msg  string
		args []interface{}

		expectOutput string
	}{
		"debug @ debug": {
			loggerLevel:  LogLevelDebug,
			messageLevel: LogLevelDebug,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [DEBUG] foo bar\n",
		},

		"info @ debug": {
			loggerLevel:  LogLevelDebug,
			messageLevel: LogLevelInfo,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [INFO]  foo bar\n",
		},

		"warn @ debug": {
			loggerLevel:  LogLevelDebug,
			messageLevel: LogLevelWarn,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [WARN]  foo bar\n",
		},

		"debug @ info": {
			loggerLevel:  LogLevelInfo,
			messageLevel: LogLevelDebug,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "",
		},

		"info @ info": {
			loggerLevel:  LogLevelInfo,
			messageLevel: LogLevelInfo,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [INFO]  foo bar\n",
		},

		"warn @ info": {
			loggerLevel:  LogLevelInfo,
			messageLevel: LogLevelWarn,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [WARN]  foo bar\n",
		},

		"debug @ warn": {
			loggerLevel:  LogLevelWarn,
			messageLevel: LogLevelDebug,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "",
		},

		"info @ warn": {
			loggerLevel:  LogLevelWarn,
			messageLevel: LogLevelInfo,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "",
		},

		"warn @ warn": {
			loggerLevel:  LogLevelWarn,
			messageLevel: LogLevelWarn,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [WARN]  foo bar\n",
		},

		"debug @ silent": {
			loggerLevel:  LogLevelSilent,
			messageLevel: LogLevelDebug,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "",
		},

		"info @ silent": {
			loggerLevel:  LogLevelSilent,
			messageLevel: LogLevelInfo,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "",
		},

		"warn @ silent": {
			loggerLevel:  LogLevelSilent,
			messageLevel: LogLevelWarn,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "",
		},

		"debug @ default": {
			loggerLevel:  0,
			messageLevel: LogLevelDebug,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "",
		},

		"info @ default": {
			loggerLevel:  0,
			messageLevel: LogLevelInfo,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [INFO]  foo bar\n",
		},

		"warn @ default": {
			loggerLevel:  0,
			messageLevel: LogLevelWarn,

			msg:  "foo %s",
			args: []interface{}{"bar"},

			expectOutput: "[memongo] [WARN]  foo bar\n",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			out := bytes.NewBufferString("")
			logger := New(log.New(out, "", 0), test.loggerLevel)

			if test.messageLevel == LogLevelDebug {
				logger.Debugf(test.msg, test.args...)
			} else if test.messageLevel == LogLevelInfo {
				logger.Infof(test.msg, test.args...)
			} else if test.messageLevel == LogLevelWarn {
				logger.Warnf(test.msg, test.args...)
			}

			got := out.String()

			assert.Equal(t, test.expectOutput, got)
		})
	}
}
