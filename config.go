package memongo

import (
	"log"
	"os"
	"path"
	"runtime"

	"github.com/benweissmann/memongo/memongolog"
	"github.com/benweissmann/memongo/mongobin"
)

// Options is the configuration options for a launched MongoDB binary
type Options struct {
	// Path to the cache for downloaded mongod binaries. Defaults to the
	// system cache location.
	CachePath string

	// If DownloadURL and MongodBin are not given, this version of MongoDB will
	// be downloaded
	MongoVersion string

	// If given, mongod will be downloaded from this URL instead of the
	// auto-detected URL based on the current platform and MongoVersion
	DownloadURL string

	// If given, this binary will be run instead of downloading a mongod binary
	MongodBin string

	// Logger for printing messages. Defaults to printing to stdout.
	Logger *log.Logger

	// A LogLevel to log at. Defaults to LogLevelInfo.
	LogLevel memongolog.LogLevel
}

func (opts *Options) fillDefaults() error {
	if opts.MongodBin == "" {
		opts.MongodBin = os.Getenv("MEMONGO_MONGOD_BIN")
	}
	if opts.MongodBin == "" {
		// The user didn't give us a local path to a binary. That means we need
		// a download URL and a cache path.

		// Determine the cache path
		if opts.CachePath == "" {
			opts.CachePath = os.Getenv("MEMONGO_CACHE_PATH")
		}
		if opts.CachePath == "" && os.Getenv("XDG_CACHE_HOME") != "" {
			opts.CachePath = path.Join(os.Getenv("XDG_CACHE_HOME"), "memongo")
		}
		if opts.CachePath == "" {
			if runtime.GOOS == "darwin" {
				opts.CachePath = path.Join(os.Getenv("HOME"), "Library", "Caches", "memongo")
			} else {
				opts.CachePath = path.Join(os.Getenv("HOME"), ".cache", "memongo")
			}
		}

		// Determine the download URL
		if opts.DownloadURL == "" {
			opts.DownloadURL = os.Getenv("MEMONGO_DOWNLOAD_URL")
		}
		if opts.DownloadURL == "" {
			spec, err := mongobin.MakeDownloadSpec(opts.MongoVersion)
			if err != nil {
				return err
			}

			opts.DownloadURL = spec.GetDownloadURL()
		}
	}

	return nil
}

func (opts *Options) getLogger() *memongolog.Logger {
	return memongolog.New(opts.Logger, opts.LogLevel)
}

func (opts *Options) getOrDownloadBinPath() (string, error) {
	if opts.MongodBin != "" {
		return opts.MongodBin, nil
	}

	// Download or fetch from cache
	binPath, err := mongobin.GetOrDownloadMongod(opts.DownloadURL, opts.CachePath, opts.getLogger())
	if err != nil {
		return "", err
	}

	return binPath, nil
}
