// Copyright 2021 Tryvium Travels LTD
// Copyright 2019-2020 Ben Weissmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
