package agentquery

import (
	"fmt"
	"os"
	"time"
)

func debugTiming(label string) func() {
	if os.Getenv("GRIMOIRE_DEBUG_TIMINGS") == "" {
		return func() {}
	}
	started := time.Now()
	return func() {
		_, _ = fmt.Fprintf(os.Stderr, "grimoire timing %s=%s\n", label, time.Since(started))
	}
}
