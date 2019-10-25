package memongo

import (
	"os"
	"testing"

	"github.com/benweissmann/memongo/mongobin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	tests := map[string]struct {
		opts *Options
		env  map[string]string
		goOS string

		// Set the Port field to -1 to expect a randomly-assigned port
		expectFilled *Options
		expectError  error
	}{
		// "minimal options, minimal env, mac": {
		// 	opts: &Options{
		// 		MongoVersion: "3.4.0",
		// 	},
		// 	env: map[string]string{
		// 		"HOME": "/home/foo",
		// 	},
		// 	goOS: "darwin",

		// 	expectFilled: &Options{
		// 		CachePath:    "/home/foo/Library/Caches/memongo",
		// 		MongoVersion: "3.4.0",
		// 		DownloadURL:  "https://fastdl.mongodb.org/osx/mongodb-osx-ssl-x86_64-3.4.0.tgz",
		// 		Port:         -1,
		// 	},
		// },
		// "minimal options, minimal env, linux": {
		// 	opts: &Options{
		// 		MongoVersion: "3.4.0",
		// 	},
		// 	env: map[string]string{
		// 		"HOME": "/home/foo",
		// 	},
		// 	goOS: "linux",

		// 	expectFilled: &Options{
		// 		CachePath:    "/home/foo/.cache/memongo",
		// 		MongoVersion: "3.4.0",
		// 		DownloadURL:  "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-3.4.0.tgz",
		// 		Port:         -1,
		// 	},
		// },
		// "explicit mongodbin":           {},
		// "env mongodbin":                {},
		// "explicit cache path":          {},
		// "env cache path":               {},
		// "xdg cache path":               {},
		// "error, no cache path":         {},
		// "explicit download URL":        {},
		// "env download url":             {},
		// "error, invalid mongo version": {},
		// "explicit port":                {},
		// "error, invalid port":          {},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			// Clear old env
			for _, e := range []string{"MEMONGO_MONGOD_BIN", "MEMONGO_CACHE_PATH", "XDG_CACHE_HOME", "HOME", "MEMONGO_DOWNLOAD_URL", "MEMONGO_MONGOD_PORT"} {
				require.NoError(t, os.Unsetenv(e))
			}

			// Set test env
			for eKey, eVal := range test.env {
				require.NoError(t, os.Setenv(eKey, eVal))
			}

			// Set test OS
			goOS = test.goOS
			mongobin.SetGOOSForTest(test.goOS)

			// Fill options
			err := test.opts.fillDefaults()

			if test.expectError == nil {
				// expect success
				assert.NoError(t, err)

				// If expected port was -1, that indicates we wanted an OS-assigned port
				assert.True(t, test.opts.Port > 1024)
				test.expectFilled.Port = test.opts.Port

				// Check equality
				assert.Equal(t, test.expectFilled, test.opts)

			} else {
				// expect error
				assert.EqualError(t, err, test.expectError.Error())
			}
		})
	}
}
