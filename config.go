package memongo

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/benweissmann/memongo/memongolog"
	"github.com/benweissmann/memongo/mongobin"
)

// Options is the configuration options for a launched MongoDB binary
type Options struct {
	// Port to run MongoDB on. If this is not specified, a random (OS-assigned)
	// port will be used
	Port int

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

	// How long to wait for mongod to start up and report a port number. Does
	// not include download time, only startup time. Defaults to 10 seconds.
	StartupTimeout time.Duration
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
			if opts.MongoVersion == "" {
				return errors.New("one of MongoVersion, DownloadURL, or MongodBin must be given")
			}
			spec, err := mongobin.MakeDownloadSpec(opts.MongoVersion)
			if err != nil {
				return err
			}

			opts.DownloadURL = spec.GetDownloadURL()
		}
	}

	// Determine the port number
	if opts.Port == 0 {
		mongoVersionEnv := os.Getenv("MEMONGO_MONGOD_PORT")
		if mongoVersionEnv != "" {
			port, err := strconv.Atoi(mongoVersionEnv)

			if err != nil {
				return fmt.Errorf("error parsing MEMONGO_MONGOD_PORT: %s", err)
			}

			opts.Port = port
		}
	}

	if opts.Port == 0 {
		// MongoDB after version 4 correctly reports what port it's running on if
		// we tell it to run on port 0, which is ideal -- we just start it on port
		// 0, the OS assigns a port, and mongo reports in the logs what port it
		// got.
		//
		// For earlier versions, mongo just print "waiting for connections on port 0"
		// which is unhelpful. So we start up a server and see what port we get,
		// then shut down that server
		if opts.MongoVersion == "" || parseMongoMajorVersion(opts.MongoVersion) < 4 {
			port, err := getFreePort()
			if err != nil {
				return fmt.Errorf("error finding a free port: %s", err)
			}

			opts.Port = port
		}

		if opts.StartupTimeout == 0 {
			opts.StartupTimeout = 10 * time.Second
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

func parseMongoMajorVersion(version string) int {
	strParts := strings.Split(version, ".")
	if len(strParts) == 0 {
		return 0
	}

	maj, err := strconv.Atoi(strParts[0])
	if err != nil {
		return 0
	}

	return maj
}

func getFreePort() (int, error) {
	// Based on: https://github.com/phayes/freeport/blob/master/freeport.go
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
