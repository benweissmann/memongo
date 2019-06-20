package monitor

import (
	"os/exec"
)

// RunMonitor runs a subprocess that kills the given child pid when the
// parent pid exits.
func RunMonitor(parent int, child int) (*exec.Cmd, error) {
	cmd := exec.Command("/bin/sh", "-c", monitorScript(parent, child))

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
