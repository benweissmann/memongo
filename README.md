# memongo

memongo is a Go package that spins up a real MongoDB server, backed by in-memory
storage, for use in testing and mocking during development. It's based on
[mongodb-memory-server](https://github.com/nodkz/mongodb-memory-server) for
NodeJS.

# Project Status

Pre-alpha. Basic tests work. Full testing and CI setup is not complete. Many features may be broken.

# Caveats and Notes

Currently, memongo only supports UNIX systems. CI will run on MacOS, Ubuntu Xenial, Ubuntu Trusty, and Ubuntu Precise. Other flavors of Linux may or may not work. CI will also run inside an Alpine Linux docker container with a system-installed copy of MongoDB.

# Basic Usage

Spin up a server for a single test:

```
func TestSomething(t *testing.T) {
  mongoServer, err := memongo.Start("4.0.5")
  if (err != nil) {
    t.Fatal(err)
  }
  defer mongoServer.Stop()

  connectAndDoStuff(mongoServer.URL(), memongo.RandomDatabase())
}
```

Spin up a server, shared between tests:

```
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
  connectAndDoStuff(mongoServer.URL(), memongo.RandomDatabase())
}
```

# Configuration

The behavior of memongo can be controlled by using
`memongo.StartWithOptions` instead of `memongo.Start`. See
[the godoc](TODO) for all the options. Many options can also be set via environment variable.

A few common use-cases are covered here:

Note that you must use MongoDB version 3.2 or greater, because the `ephemeralForTest` storage engine was not present before 3.2.

## Set the cache path

memongo downloads a pre-compiled binary of MongoDB from https://www.mongodb.org and caches it on your local system. This path is set by (in order of preference):

- The `CachePath` passed to `memongo.StartWithOptions`
- The environment variable `MEMONGO_CACHE_PATH`
- If `XDG_CACHE_HOME` is set, `$XDG_CACHE_HOME/memongo`
- `~/.cache/memongo` on Linux, or `~/Library/Caches/memongo` on MacOS

## Override download URL

By default, memongo tries to detect the platform you're running on and download an official MongoDB release for it. If memongo doesn't yet support your platform, of you'd like to use a custom version of MongoDB, you can pass `DownloadURL` to `memongo.StartWithOptions` or set the environment variable `MEMONGO_DOWNLOAD_URL`.

memongo's caching will still work with custom download URLs.

## Use a custom MongoDB binary

If you'd like to bypass memongo's download beahvior entirely, you can pass `MongodBin` to `memongo.StartWithOptions`, or set the environment variable `MEMONGO_MONGOD_BIN` to the path to a `mongod` binary. memongo will use this binary instead of downloading one.

If you're running on a platform that doesn't have an official MongoDB release (such as Alpine), you'll need to use this option.

## Reduce or increase logging

By default, memongo logs at an "info" level. You may call `StartWithOptions` with `LogLevel: memongo.LogLevelWarn` for fewer logs, `LogLevel: memongo.LogLevelSilent` for no logs, or `LogLevel: memongo.LogLevelDebug` for verbose logs (including full logs from MongoDB).

By default, memongo logs to stdout. To log somewhere else, specify a `Logger` in `StartWithOptions`.
