package monitor

import (
	"fmt"
	"os/exec"
)

// Run starts a subprocess that kills the given child pid when the
// parent pid exits.
func Run(parent int, child int) (*exec.Cmd, error) {
	cmd := exec.Command("/bin/sh", "-c", monitorScript(parent, child))

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func monitorScript(parent int, child int) string {
	return fmt.Sprintf(
		"while ps -o pid= -p %d; "+
			"do sleep 1; "+
			"done; "+
			"kill -9 %d",
		parent, child)
}
