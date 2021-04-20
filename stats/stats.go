package stats

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"time"
)

// Stat bundles the metrics we collect into a single struct
type Stat struct {
	SentBytes uint64
	ReceivedBytes uint64
}

// statPipe is a channel used to
var stats Stat

// PushStat pushes the given stat
func PushStat(s Stat) {
	stats.SentBytes += s.SentBytes
	stats.ReceivedBytes += s.ReceivedBytes
}

// PrintStats periodically prints the collected statistics
func PrintStats() {
	t := time.NewTicker(2 * time.Second)

	for {
		// Wait until we get a tick
		<-t.C

		// Format numbers and write them out
		bIn := humanize.Bytes(stats.ReceivedBytes)
		bOut := humanize.Bytes(stats.SentBytes)

		fmt.Printf("%s in / %s out\r", bIn, bOut)
	}
}