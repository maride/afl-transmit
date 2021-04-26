package stats

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"time"
)

// Stat bundles the metrics we collect into a single struct
type Stat struct {
	SentBytes uint64
	ReceivedBytes uint64
	RegisteredPeers uint8
	AlivePeer uint8
}

// statPipe is a channel used to
var stats Stat

// printStats sets whether we should print stats or not
var printStats bool

// RegisterStatsFlags registers all flags required by the stats module
func RegisterStatsFlags() {
	flag.BoolVar(&printStats, "print-stats", true, "Print traffic statistics every few seconds")
}

// PushStat pushes the given stat
// Note that SentBytes, ReceivedBytes and RegisteredPeers are added to the current number,
// while AlivePeer is interfaced with SetAlivePeers and is left ignored by PushStat
func PushStat(s Stat) {
	stats.SentBytes += s.SentBytes
	stats.ReceivedBytes += s.ReceivedBytes
	stats.RegisteredPeers += s.RegisteredPeers
}

// SetAlivePeers sets the number of alive peers, means peers we could connect to
func SetAlivePeers(n uint8) {
	stats.AlivePeer = n
}

// PrintStats periodically prints the collected statistics
func PrintStats() {
	// Check if we should print stats
	if !printStats {
		return
	}

	t := time.NewTicker(2 * time.Second)

	for {
		// Wait until we get a tick
		<-t.C

		// Format numbers and write them out
		bIn := humanize.Bytes(stats.ReceivedBytes)
		bOut := humanize.Bytes(stats.SentBytes)

		fmt.Printf("Traffic: %s in / %s out | Peers: %d seen / %d registered\t\r", bIn, bOut, stats.AlivePeer, stats.RegisteredPeers)
	}
}