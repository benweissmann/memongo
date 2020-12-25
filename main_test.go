package memongo

import (
	"context"
	"testing"
	"time"

	"github.com/benweissmann/memongo/memongolog"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestDefaultOptions(t *testing.T) {
	versions := []string{"3.2.22", "3.4.21", "3.6.13", "4.0.13", "4.2.1"}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			server, err := StartWithOptions(&Options{
				MongoVersion: version,
				LogLevel:     memongolog.LogLevelDebug,
			})
			require.NoError(t, err)
			defer server.Stop()

			client, err := mongo.NewClient(options.Client().ApplyURI(server.URI()))

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = client.Connect(ctx)
			if err != nil {
				server.logger.Warnf("err Connect: %v", err)
			}

			require.NoError(t, err)
			require.NoError(t, client.Ping(context.Background(), readpref.Primary()))
		})
	}
}

func TestWithReplica(t *testing.T) {
	versions := []string{"3.6.13", "4.0.13", "4.2.1"}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			server, err := StartWithOptions(&Options{
				MongoVersion:     version,
				LogLevel:         memongolog.LogLevelDebug,
				ShouldUseReplica: true,
			})
			require.NoError(t, err)
			defer server.Stop()

			client, err := mongo.NewClient(options.Client().ApplyURI(server.URI()))

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = client.Connect(ctx)
			if err != nil {
				server.logger.Warnf("err Connect: %v", err)
			}

			require.NoError(t, err)
			require.NoError(t, client.Ping(context.Background(), readpref.Primary()))
		})
	}
}
