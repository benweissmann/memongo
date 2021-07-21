# memongo

[![CI](https://github.com/tryvium-travels/memongo/workflows/Go/badge.svg)](https://github.com/tryvium-travels/memongo/actions?workflow=Go) [![GoDoc](https://godoc.org/github.com/tryvium-travels/memongo?status.svg)](https://godoc.org/github.com/tryvium-travels/memongo) [![Go Report Card](https://goreportcard.com/badge/github.com/tryvium-travels/memongo)](https://goreportcard.com/report/github.com/tryvium-travels/memongo)

`memongo` is a Go package that spins up a real MongoDB server, backed by in-memory
storage, for use in testing and mocking during development. It's based on
[mongodb-memory-server](https://github.com/nodkz/mongodb-memory-server) for
NodeJS.

In general, it's better to mock out interaction with the database, so you don't
need to run a Mongo server during testing. But becuase most Mongo clients use
a fluent interface that's tough to mock, and sometimes you need to test the
queries themselves, it's often helpful to be able to spin up a Mongo server
quickly and easily. That's where `memongo` comes in!

# Project Status

Beta. Tests and CI are set up and working, but more esoteric configurations may not work. If Memongo isn't working on your platform, you might want to use `memongo.StartWithOptions()` and pass the correct download URL for your platform manually.

# Caveats and Notes

Currently, `memongo` only supports UNIX systems. CI will run on MacOS, Ubuntu Xenial, Ubuntu Trusty, and Ubuntu Precise. Other flavors of Linux may or may not work.

# Basic Usage

Spin up a server for a single test:

```go
func TestSomething(t *testing.T) {
  mongoServer, err := memongo.Start("4.0.5")
  if (err != nil) {
    t.Fatal(err)
  }
  defer mongoServer.Stop()

  connectAndDoStuff(mongoServer.URI(), memongo.RandomDatabase())
}
```

Spin up a server, shared between tests:

```go
var mongoServer memongo.Server;

func TestMain(m *testing.M) {
  mongoServer, err = memongo.Start("4.0.5")
  if (err != nil) {
    log.Fatal(err)
  }
  defer mongoServer.Stop()

  os.Exit(m.Run())
}

func TestSomething(t *testing.T) {
  connectAndDoStuff(mongoServer.URI(), memongo.RandomDatabase())
}
```

# How it works

Behind the scenes, when you run `Start()`, a few things are happening:

1. If you specified a MongoDB version number (rather than a URL or binary path),
   `memongo` detects your operating system and platform to determine the
   download URL for the right MongoDB binary.

2. If you specified a MongoDB version number or download URL, `memongo`
   downloads MongoDB to a cache location. For future runs, `memongo` will just
   use the copy from the cache. You only need to be connected to the internet
   the first time you run `Start()` for a particular MongoDB version.

3. `memongo` starts a process running the downloaded `mongod` binary. It uses
   the `ephemeralForTest` storage engine, a temporary directory for a `dbpath`,
   and a random free port number.

4. `memongo` also starts up a "watcher" process. This process is a simple
   portable shell script that kills the `mongod` process when the current
   process exits. This ensures that we don't leave behind `mongod` processes,
   even if your tests exit uncleanly or you don't call `Stop()`.

# Configuration

The behavior of `memongo` can be controlled by using
`memongo.StartWithOptions` instead of `memongo.Start`. See
[the godoc](https://godoc.org/github.com/tryvium-travels/memongo) for all the options. Many options can also be set via environment variable.

A few common use-cases are covered here:

Note that you must use MongoDB version 3.2 or greater, because the `ephemeralForTest` storage engine was not present before 3.2.

## Set the cache path

`memongo` downloads a pre-compiled binary of MongoDB from https://www.mongodb.org and caches it on your local system. This path is set by (in order of preference):

- The `CachePath` passed to `memongo.StartWithOptions`
- The environment variable `MEMONGO_CACHE_PATH`
- If `XDG_CACHE_HOME` is set, `$XDG_CACHE_HOME/memongo`
- `~/.cache/memongo` on Linux, or `~/Library/Caches/memongo` on MacOS

## Override download URL

By default, `memongo` tries to detect the platform you're running on and download an official MongoDB release for it. If `memongo` doesn't yet support your platform, of you'd like to use a custom version of MongoDB, you can pass `DownloadURL` to `memongo.StartWithOptions` or set the environment variable `MEMONGO_DOWNLOAD_URL`.

`memongo`'s caching will still work with custom download URLs.

## Use a custom MongoDB binary

If you'd like to bypass `memongo`'s download beahvior entirely, you can pass `MongodBin` to `memongo.StartWithOptions`, or set the environment variable `MEMONGO_MONGOD_BIN` to the path to a `mongod` binary. `memongo` will use this binary instead of downloading one.

If you're running on a platform that doesn't have an official MongoDB release (such as Alpine), you'll need to use this option.

## Reduce or increase logging

By default, `memongo` logs at an "info" level. You may call `StartWithOptions` with `LogLevel: memongolog.LogLevelWarn` for fewer logs, `LogLevel: memongolog.LogLevelSilent` for no logs, or `LogLevel: memongolog.LogLevelDebug` for verbose logs (including full logs from MongoDB).

By default, `memongo` logs to stdout. To log somewhere else, specify a `Logger` in `StartWithOptions`.
