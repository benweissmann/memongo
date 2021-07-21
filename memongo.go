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

package memongo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tryvium-travels/memongo/memongolog"
	"github.com/tryvium-travels/memongo/monitor"
)

// Server represents a running MongoDB server
type Server struct {
	cmd        *exec.Cmd
	watcherCmd *exec.Cmd
	dbDir      string
	logger     *memongolog.Logger
	port       int
}

// Start runs a MongoDB server at a given MongoDB version using default options
// and returns the Server.
func Start(version string) (*Server, error) {
	return StartWithOptions(&Options{
		MongoVersion: version,
	})
}

// StartWithOptions is like Start(), but accepts options.
func StartWithOptions(opts *Options) (*Server, error) {
	err := opts.fillDefaults()
	if err != nil {
		return nil, err
	}

	logger := opts.getLogger()

	logger.Infof("Starting MongoDB with options %#v", opts)

	binPath, err := opts.getOrDownloadBinPath()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Using binary %s", binPath)

	// Create a db dir. Even the ephemeralForTest engine needs a dbpath.
	dbDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	// Construct the command and attach stdout/stderr handlers

	//  Safe to pass binPath and dbDir
	//nolint:gosec
	cmd := exec.Command(binPath, "--storageEngine", "ephemeralForTest", "--dbpath", dbDir, "--port", strconv.Itoa(opts.Port))

	stdoutHandler, startupErrCh, startupPortCh := stdoutHandler(logger)
	cmd.Stdout = stdoutHandler
	cmd.Stderr = stderrHandler(logger)

	logger.Debugf("Starting mongod")

	// Run the server
	err = cmd.Start()
	if err != nil {
		remErr := os.RemoveAll(dbDir)
		if remErr != nil {
			logger.Warnf("error removing data directory: %s", remErr)
		}

		return nil, err
	}

	logger.Debugf("Started mongod; starting watcher")

	// Start a watcher: the watcher is a subprocess that ensure if this process
	// dies, the mongo server will be killed (and not reparented under init)
	watcherCmd, err := monitor.RunMonitor(os.Getpid(), cmd.Process.Pid)
	if err != nil {
		killErr := cmd.Process.Kill()
		if killErr != nil {
			logger.Warnf("error stopping mongo process: %s", killErr)
		}

		remErr := os.RemoveAll(dbDir)
		if remErr != nil {
			logger.Warnf("error removing data directory: %s", remErr)
		}

		return nil, err
	}

	logger.Debugf("Started watcher; waiting for mongod to report port number")
	startupTime := time.Now()

	// Wait for the stdout handler to report the server's port number (or a
	// startup error)
	var port int
	select {
	case p := <-startupPortCh:
		port = p
	case err := <-startupErrCh:
		killErr := cmd.Process.Kill()
		if killErr != nil {
			logger.Warnf("error stopping mongo process: %s", killErr)
		}

		remErr := os.RemoveAll(dbDir)
		if remErr != nil {
			logger.Warnf("error removing data directory: %s", remErr)
		}

		return nil, err
	case <-time.After(opts.StartupTimeout):
		killErr := cmd.Process.Kill()
		if killErr != nil {
			logger.Warnf("error stopping mongo process: %s", killErr)
		}

		remErr := os.RemoveAll(dbDir)
		if remErr != nil {
			logger.Warnf("error removing data directory: %s", remErr)
		}

		return nil, errors.New("timed out waiting for mongod to start")
	}

	logger.Debugf("mongod started up and reported a port number after %s", time.Since(startupTime).String())

	// Return a Memongo server
	return &Server{
		cmd:        cmd,
		watcherCmd: watcherCmd,
		dbDir:      dbDir,
		logger:     logger,
		port:       port,
	}, nil
}

// Port returns the port the server is listening on.
func (s *Server) Port() int {
	return s.port
}

// URI returns a mongodb:// URI to connect to
func (s *Server) URI() string {
	return fmt.Sprintf("mongodb://localhost:%d", s.port)
}

// URIWithRandomDB returns a mongodb:// URI to connect to, with
// a random database name (e.g. mongodb://localhost:1234/somerandomname)
func (s *Server) URIWithRandomDB() string {
	return fmt.Sprintf("mongodb://localhost:%d/%s", s.port, RandomDatabase())
}

// Stop kills the mongo server
func (s *Server) Stop() {
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
var reReady = regexp.MustCompile(`waiting for connections.*port\D*(\d+)`)
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
func stdoutHandler(log *memongolog.Logger) (io.Writer, <-chan error, <-chan int) {
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
func stderrHandler(log *memongolog.Logger) io.Writer {
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
