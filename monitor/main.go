package monitor

import (
	"fmt"
	"os/exec"
)

// RunMonitor runs a subprocess that kills the given child pid when the
// parent pid exits.
func RunMonitor(parent int, child int) (*exec.Cmd, error) {
	// monitorScript returns a safe script; it's parameterized only by integers
	//nolint:gosec
	cmd := exec.Command("/bin/sh", "-c", monitorScript(parent, child))

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Error starting watcher process: %s", err)
	}

	return cmd, nil
}
