# dp-mongodb-in-memory
Library that runs an in-memory MongoDB instance for Go unit tests.

## How it works
- It detects your operating system and platform to determine the download URL for the right MongoDB binary.

- It will download MongoDB and store it in a [cache location](#cache-location). Any following execution will use the copy from the cache. Therefore internet connection is only required the first time a particular MongoDB version is used.

- It will start a process running the downloaded `mongod` binary. It uses the `ephemeralForTest` storage engine, a temporary directory for a `dbpath` and a random free port number.

- Additionally, a _watcher_ process will start in background ensuring that the mongod process is killed when the current process exits. This guarantees that no process is left behind even if the tests exit uncleanly or you don't call `Stop()`.

### Supported versions
The following Unix systems are supported:
- MacOS
- Ubuntu 16.04 or greater
- Debian 9.2 or greater

The supported MongoDB versions are 4.4 and above.

### Cache location

The downloaded mongodb binary will be stored in a local cache: a folder named `dp-mongodb-in-memory` living on the machine base cache directory. That is `$XDG_CACHE_HOME` if such environment variable is set or `~/.cache` (Linux) and `~/Library/Caches` (MacOS) if not.


## Installation

To install this package run:

```bash
go get github.com/ONSdigital/dp-mongodb-in-memory
```

## Usage

Simply call `Start()` with the MongoDB version you want to use and it will spin up a server for test. You can then use `URI()` to connect a client to it. 
Call `Stop()` when you are done with the server.

```go
package example

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mim "github.com/ONSdigital/dp-mongodb-in-memory"
)

func TestExample(t *testing.T) {
	server, err := mim.Start("5.0.2")
	if err != nil {
		// Deal with error
	}
	defer server.Stop()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(server.URI()))
	if err != nil {
		// Deal with error
	}

	//Use client as needed
	client.Ping(context.Background(), nil)
}

```

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
