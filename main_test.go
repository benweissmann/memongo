package mim

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestStart(t *testing.T) {
	versions := []string{"4.4.8", "5.0.2"}

	for _, version := range versions {
		Convey("Given the version "+version, t, func() {
			Convey("When the Start method is called", func() {
				server, err := Start(version)
				defer server.Stop()

				Convey("Then no error is returned", func() {
					So(err, ShouldBeNil)
					Convey("And the mongod process has run", func() {
						So(server, ShouldNotBeNil)
						So(server.cmd, ShouldNotBeNil)
						So(server.dbDir, ShouldNotBeBlank)
						So(server.port, ShouldNotBeEmpty)
						So(server.cmd.Args[0], ShouldEndWith, "mongod")
						So(server.cmd.Args[1], ShouldEqual, "--storageEngine")
						So(server.cmd.Args[2], ShouldEqual, "ephemeralForTest")
						So(server.cmd.Args[3], ShouldEqual, "--dbpath")
						So(server.cmd.Args[4], ShouldEqual, server.dbDir)
						So(server.cmd.Args[5], ShouldEqual, "--port")
						So(server.cmd.Args[6], ShouldEqual, "0")
					})
					Convey("And the watcher process has run", func() {
						expectedScript := fmt.Sprintf("while ps -o pid= -p %d; "+
							"do sleep 1; "+
							"done; "+
							"kill -9 %d",
							os.Getpid(), server.cmd.Process.Pid)

						So(server, ShouldNotBeNil)
						So(server.watcherCmd, ShouldNotBeNil)
						So(server.watcherCmd.Args[0], ShouldEndWith, "/bin/sh")
						So(server.watcherCmd.Args[1], ShouldEqual, "-c")
						So(server.watcherCmd.Args[2], ShouldEqual, expectedScript)
					})
					Convey("And the server accepts connections", func() {
						client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(server.URI()))
						So(err, ShouldBeNil)
						So(client, ShouldNotBeNil)
						So(client.Ping(context.Background(), nil), ShouldBeNil)
					})
				})
			})
		})
	}
}
