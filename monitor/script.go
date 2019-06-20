package monitor

import "fmt"

func monitorScript(parent int, child int) string {
	return fmt.Sprintf(
		"while kill -0 %d; do "+
			"sleep 1; "+
			"done; "+
			"kill -9 %d ",
		parent, child)
}
