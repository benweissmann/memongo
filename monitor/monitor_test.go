package monitor

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRun(t *testing.T) {
	Convey("Given a parent and child pid", t, func() {
		parentPid := 88
		childPid := 217
		Convey("When the Run method is called", func() {
			cmd, err := Run(parentPid, childPid)
			defer cmd.Process.Kill()
			Convey("Then the right script command has run", func() {
				expectedScript := fmt.Sprintf("while ps -o pid= -p %d; "+
					"do sleep 1; "+
					"done; "+
					"kill -9 %d",
					parentPid, childPid)

				So(err, ShouldBeNil)
				So(cmd.Args[0], ShouldEndWith, "/bin/sh")
				So(cmd.Args[1], ShouldEqual, "-c")
				So(cmd.Args[2], ShouldEqual, expectedScript)
			})
		})
	})
}
