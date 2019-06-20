package memongo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/benweissmann/memongo/mongobin"
	"github.com/benweissmann/memongo/monitor"
)

type MemongoOptions struct {
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
	LogLevel LogLevel
}

type MemongoServer struct {
	cmd        *exec.Cmd
	watcherCmd *exec.Cmd
	dbDir      string
	logger     *logger
	port       int
}

func Start(version string) (*MemongoServer, error) {
	return StartWithOptions(&MemongoOptions{
		MongoVersion: version,
	})
}

func StartWithOptions(opts *MemongoOptions) (*MemongoServer, error) {
	// Get a logger
	logger := newLogger(opts.Logger, opts.LogLevel)

	// Download if needed
	binPath := opts.MongodBin
	if binPath == "" {
		binPath = os.Getenv("MEMONGO_MONGOD_BIN")
	}
	if binPath == "" {
		// Determine the cache path
		cachePath := opts.CachePath
		if cachePath == "" {
			cachePath = os.Getenv("MEMONGO_CACHE_PATH")
		}
		if cachePath == "" && os.Getenv("XDG_CACHE_HOME") != "" {
			cachePath = path.Join(os.Getenv("XDG_CACHE_HOME"), "/memongo")
		}
		if cachePath == "" {
			user, err := user.Current()
			if err != nil {
				return nil, fmt.Errorf("unable to get current user: %s", err)
			}

			if runtime.GOOS == "darwin" {
				cachePath = path.Join(user.HomeDir, "Library", "Caches", "memongo")
			} else {
				cachePath = path.Join(user.HomeDir, ".cache", "memongo")
			}
		}

		// Determine the download URL
		downloadURL := opts.DownloadURL
		if downloadURL == "" {
			downloadURL = os.Getenv("MEMONGO_DOWNLOAD_URL")
		}
		if downloadURL == "" {
			spec, err := mongobin.MakeDownloadSpec(opts.MongoVersion)
			if err != nil {
				return nil, err
			}

			downloadURL = spec.GetDownloadURL()
		}

		// Download or fetch from cache
		var err error

		downloadLogger := log.New(ioutil.Discard, "", 0)
		if opts.LogLevel <= LogLevelInfo {
			downloadLogger = logger.out
		}

		binPath, err = mongobin.GetOrDownloadMongod(downloadURL, cachePath, downloadLogger)
		if err != nil {
			return nil, err
		}
	}

	// Create a db dir. Even the ephemeralForTest engine needs a dbpath.
	dbDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	// Construct the command and attach stdout/stderr handlers
	cmd := exec.Command(binPath, "--storageEngine", "ephemeralForTest", "--dbpath", dbDir, "--port", "0")

	stdoutHandler, startupErrCh, startupPortCh := stdoutHandler(logger)
	cmd.Stdout = stdoutHandler
	cmd.Stderr = stderrHandler(logger)

	// Run the server
	err = cmd.Start()
	if err != nil {
		os.RemoveAll(dbDir)
		return nil, err
	}

	// Start a watcher: the watcher is a subprocess that ensure if this process
	// dies, the mongo server will be killed (and not reparented under init)
	watcherCmd, err := monitor.RunMonitor(os.Getpid(), cmd.Process.Pid)
	if err != nil {
		_ = cmd.Process.Kill()
		os.RemoveAll(dbDir)
		return nil, err
	}

	// Wait for the stdout handler to report the server's port number (or a
	// startup error)
	var port int
	select {
	case p := <-startupPortCh:
		port = p
	case err := <-startupErrCh:
		_ = cmd.Process.Kill()
		os.RemoveAll(dbDir)
		return nil, err
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		os.RemoveAll(dbDir)
		return nil, errors.New("timed out waiting for mongod to start")
	}

	// Return a Memongo server
	return &MemongoServer{
		cmd:        cmd,
		watcherCmd: watcherCmd,
		dbDir:      dbDir,
		logger:     logger,
		port:       port,
	}, nil
}

// Port returns the port the server is listening on.
func (s *MemongoServer) Port() int {
	return s.port
}

// URI returns a mongodb:// URI to connect to
func (s *MemongoServer) URI() string {
	return "mongodb://localhost:" + strconv.Itoa(s.port)
}

// Stop kills the mongo server
func (s *MemongoServer) Stop() {
	err := s.cmd.Process.Kill()
	if err != nil {
		s.logger.Warnf("error stopping mongod process: %s", err)
		return
	}

	err = s.watcherCmd.Process.Kill()
	if err != nil {
		s.logger.Warnf("error stopping watcher process: %s", err)
		return
	}

	err = os.RemoveAll(s.dbDir)
	if err != nil {
		s.logger.Warnf("error removing data directory: %s", err)
		return
	}
}

// Cribbed from https://github.com/nodkz/mongodb-memory-server/blob/master/packages/mongodb-memory-server-core/src/util/MongoInstance.ts#L206
var reReady = regexp.MustCompile(`waiting for connections on port (\d+)`)
var reAlreadyInUse = regexp.MustCompile("addr already in use")
var reAlreadyRunning = regexp.MustCompile("mongod already running")
var rePermissionDenied = regexp.MustCompile("mongod permission denied")
var reDataDirectoryNotFound = regexp.MustCompile("data directory .*? not found")
var reShuttingDown = regexp.MustCompile("shutting down with code")

// The stdout handler relays lines from mongod's stout to our logger, and also
// watches during startup for error or success messages.
//
// It returns two channels: an error channel and a port channel. Only one
// message will be sent to one of these two channels. A port number will
// be sent to the port channel if the server start up correctly, and an
// error will be send to the error channel if the server does not start up
// correctly.
func stdoutHandler(log *logger) (io.Writer, <-chan error, <-chan int) {
	errChan := make(chan error)
	portChan := make(chan int)

	reader, writer := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(reader)
		haveSentMessage := false

		for scanner.Scan() {
			line := scanner.Text()

			log.Debugf("[Mongod stdout] %s", line)

			if !haveSentMessage {
				downcaseLine := strings.ToLower(line)

				if match := reReady.FindStringSubmatch(downcaseLine); match != nil {
					port, err := strconv.Atoi(match[1])
					if err != nil {
						errChan <- errors.New("Could not parse port from mongod log line: " + downcaseLine)
					} else {
						portChan <- port
					}
					haveSentMessage = true
				} else if reAlreadyInUse.MatchString(downcaseLine) {
					errChan <- errors.New("Mongod startup failed, address in use")
					haveSentMessage = true
				} else if reAlreadyRunning.MatchString(downcaseLine) {
					errChan <- errors.New("Mongod startup failed, already running")
					haveSentMessage = true
				} else if rePermissionDenied.MatchString(downcaseLine) {
					errChan <- errors.New("mongod startup failed, permission denied")
					haveSentMessage = true
				} else if reDataDirectoryNotFound.MatchString(downcaseLine) {
					errChan <- errors.New("Mongod startup failed, data directory not found")
					haveSentMessage = true
				} else if reShuttingDown.MatchString(downcaseLine) {
					errChan <- errors.New("Mongod startup failed, server shut down")
					haveSentMessage = true
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Warnf("reading mongod stdin failed: %s", err)
		}

		if !haveSentMessage {
			errChan <- errors.New("Mongod exited before startup completed")
		}
	}()

	return writer, errChan, portChan
}

// The stderr handler just relays messages from stderr to our logger
func stderrHandler(log *logger) io.Writer {
	reader, writer := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			log.Debugf("[Mongod stderr] %s", scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Warnf("reading mongod stdin failed: %s", err)
		}
	}()

	return writer
}
