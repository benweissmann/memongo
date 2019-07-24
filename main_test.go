package memongo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestBasic(t *testing.T) {
	server, err := Start("4.0.5")
	require.NoError(t, err)
	defer server.Stop()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(server.URI()))
	require.NoError(t, err)

	require.NoError(t, client.Ping(context.Background(), nil))
}
