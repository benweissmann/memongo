package mim

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/ONSdigital/dp-mongodb-in-memory/download"
	"github.com/ONSdigital/dp-mongodb-in-memory/monitor"
	"github.com/ONSdigital/log.go/v2/log"
)

// max time allowed for mongo to start
const timeout = 5 * time.Second

// Server represents a running MongoDB server.
type Server struct {
	cmd        *exec.Cmd
	watcherCmd *exec.Cmd
	dbDir      string
	port       int
}

func init() {
	log.Namespace = "dp-mongodb-in-memory"
}

// Start runs a MongoDB server at a given version using a random free port
// and returns the Server.
func Start(version string) (*Server, error) {
	server := new(Server)

	binPath, err := getOrDownloadBinPath(version)
	if err != nil {
		log.Fatal(context.Background(), "Could not find mongodb", err)
		return nil, err
	}

	// Create a db dir. Even the ephemeralForTest engine needs a dbpath.
	server.dbDir, err = ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(context.Background(), "Error creating data directory", err)
		return nil, err
	}

	log.Info(context.Background(), "Starting mongod server", log.Data{"binPath": binPath, "dbDir": server.dbDir})
	// By specifying port 0, mongo will find and use an available port
	server.cmd = exec.Command(binPath, "--storageEngine", "ephemeralForTest", "--dbpath", server.dbDir, "--port", "0")

	startupErrCh := make(chan error)
	startupPortCh := make(chan int)
	stdHandler := stdHandler(startupPortCh, startupErrCh)
	server.cmd.Stdout = stdHandler
	server.cmd.Stderr = stdHandler

	// Run the server
	err = server.cmd.Start()
	if err != nil {
		log.Fatal(context.Background(), "Could not start mongodb", err)
		server.Stop()
		return nil, err
	}

	log.Info(context.Background(), "Starting watcher")
	// Start a watcher: the watcher is a subprocess that ensures if this process
	// dies, the mongo server will be killed (and not reparented under init)
	server.watcherCmd, err = monitor.Run(os.Getpid(), server.cmd.Process.Pid)
	if err != nil {
		log.Error(context.Background(), "Could not start watcher", err)
		server.Stop()
		return nil, err
	}

	select {
	case server.port = <-startupPortCh:
	case err := <-startupErrCh:
		server.Stop()
		return nil, err
	case <-time.After(timeout):
		server.Stop()
		return nil, errors.New("timed out waiting for mongod to start")
	}

	log.Info(context.Background(), fmt.Sprintf("mongod started up and reported port number %d", server.port))

	return server, nil
}

// Stop kills the mongo server.
func (s *Server) Stop() {
	if s.cmd != nil {
		err := s.cmd.Process.Kill()
		if err != nil {
			log.Error(context.Background(), "Error stopping mongod process", err, log.Data{"pid": s.cmd.Process.Pid})
		}
	}

	if s.watcherCmd != nil {
		err := s.watcherCmd.Process.Kill()
		if err != nil {
			log.Error(context.Background(), "error stopping watcher process", err, log.Data{"pid": s.watcherCmd.Process.Pid})
		}
	}

	err := os.RemoveAll(s.dbDir)
	if err != nil {
		log.Error(context.Background(), "Error removing data directory", err, log.Data{"dir": s.dbDir})
	}
}

// Port returns the port the server is listening on.
func (s *Server) Port() int {
	return s.port
}

// URI returns a mongodb:// URI to connect to
func (s *Server) URI() string {
	return fmt.Sprintf("mongodb://localhost:%d", s.port)
}

func getOrDownloadBinPath(version string) (string, error) {
	config, err := download.NewConfig(version)
	if err != nil {
		log.Error(context.Background(), "Failed to create config", err)
		return "", err
	}

	if err := download.GetMongoDB(*config); err != nil {
		return "", err
	}
	return config.MongoPath(), nil
}

// stdHandler handler relays messages from stdout/stderr to our logger.
// It accepts 2 channels:
// errCh will receive any error logged,
// okCh will receive the port number if mongodb started successfully
func stdHandler(okCh chan<- int, errCh chan<- error) io.Writer {
	reader, writer := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			text := scanner.Text()
			var logMessage log.Data
			err := json.Unmarshal([]byte(text), &logMessage)
			if err != nil {
				// Output the message as is if not json
				log.Info(context.Background(), fmt.Sprintf("[mongod] %s", text))
			} else {
				message := logMessage["msg"]
				severity := logMessage["s"]
				if severity == "E" || severity == "F" {
					// error or fatal
					errCh <- fmt.Errorf("Mongod startup failed: %s", message)
				} else if severity == "I" && message == "Waiting for connections" {
					// Mongo running successfully: find port
					attr := logMessage["attr"].(map[string]interface{})
					okCh <- int(attr["port"].(float64))
				}

				log.Info(context.Background(), fmt.Sprintf("[mongod] %s", message), logMessage)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Error(context.Background(), "reading mongod stdout/stderr failed: %s", err)
		}
	}()

	return writer
}
