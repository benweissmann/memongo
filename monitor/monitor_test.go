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
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runSleep() *os.Process {
	cmd := exec.Command("/bin/sh", "sleep", "10")

	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	return cmd.Process
}

func TestMonitor(t *testing.T) {
	parent := runSleep()
	child := runSleep()

	// Start the monitor
	_, err := RunMonitor(parent.Pid, child.Pid)
	require.NoError(t, err)

	// Kill the parent
	require.NoError(t, parent.Kill())

	// Child should die within 3 seconds
	startWait := time.Now()
	_, err = child.Wait()
	require.NoError(t, err)

	assert.True(t, time.Since(startWait).Seconds() < 3)
}
