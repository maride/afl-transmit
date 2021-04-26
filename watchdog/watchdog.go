package watchdog

import (
	"flag"
	"fmt"
	"github.com/maride/afl-transmit/logistic"
	"github.com/maride/afl-transmit/net"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	rescan int
	isMainNode bool
)

// RegisterWatchdogFlags registers required flags for the watchdog
func RegisterWatchdogFlags() {
	flag.IntVar(&rescan, "rescan", 30, "Minutes to wait before rescanning local fuzzer directories")
	flag.BoolVar(&isMainNode, "main", false, "Set this option if this afl-transmit instance is running on the node running the main afl-fuzz instance")
}

// Watch over the specified directory, send updates to peers and re-scan after the specified amount of seconds
func WatchFuzzers(outputDirectory string) {
	// Loop forever
	for {
		// Pack important parts of the fuzzers into an archive
		packedFuzzers, packerErr := logistic.PackFuzzers(getTargetFuzzers(outputDirectory), outputDirectory)
		if packerErr != nil {
			log.Printf("Failed to pack fuzzer: %s", packerErr)
			continue
		}

		// and send it to our peers
		go net.SendToPeers(packedFuzzers)

		// Sleep a bit
		time.Sleep(time.Duration(rescan) * time.Minute)
	}
}

// Searches in the specified output directory for fuzzers which we want to pack.
// Identifying the main fuzzer is done by searching for the file "is_main_node"
// Identifying local secondary fuzzers is done by searching for a file which is not shared between nodes as it is not required for clusterized fuzzing: 'cmdline'.
// Please note that this is not failsafe, e.g. if you already synced your fuzzers over a different tool or by hand before.
//
// - If we are running in main mode (--main), we want to share the main fuzzers' queue with the secondary nodes
// - If we are not running in main mode, we want to share all of our fuzzers' queues with the main node
//
// (see AFLplusplus/src/afl-fuzz.run.c:542)
func getTargetFuzzers(outputDirectory string) []string {
	if isMainNode {
		// find main fuzzer directory - its the only one required by the secondaries

		// List files (read: fuzzers) in output directory
		filesInDir, readErr := ioutil.ReadDir(outputDirectory)
		if readErr != nil {
			log.Printf("Failed to list directory content of %s: %s", outputDirectory, readErr)
			return nil
		}

		for _, f := range filesInDir {
			// Get stat for maybe-existent file
			mainnodePath := fmt.Sprintf("%s%c%s%cis_main_node", outputDirectory, os.PathSeparator, f.Name(), os.PathSeparator)
			_, statErr := os.Stat(mainnodePath)
			if os.IsNotExist(statErr) {
				// File does not exist. Not the ~~chosen one~~ main fuzzer. Next.
				continue
			} else if statErr != nil {
				// An error occurred. File is maybe in a Schrödinger state.
				log.Printf("Unable to stat file %s: %s", mainnodePath, statErr)
				continue
			}

			// is_main_node file exists, so we found the correct fuzzer directory - return it
			fullPath := fmt.Sprintf("%s%c%s", outputDirectory, os.PathSeparator, f.Name())
			return []string{
				fullPath,
			}
		}

		// Failed to find the main node - probably we are in --main mode by accident
		log.Printf("Unable to find main node in %s - sure afl-transmit should run in --main mode?", outputDirectory)
		return nil
	} else {
		// get all fuzzer directories which are locally available
		var localFuzzers []string

		// List files (read: fuzzers) in output directory
		filesInDir, readErr := ioutil.ReadDir(outputDirectory)
		if readErr != nil {
			log.Printf("Failed to list directory content of %s: %s", outputDirectory, readErr)
			return nil
		}

		// Walk over each fuzzer and search for 'cmdline' file
		for _, f := range filesInDir {
			// Get stat for maybe-existent file
			cmdlinePath := fmt.Sprintf("%s%c%s%ccmdline", outputDirectory, os.PathSeparator, f.Name(), os.PathSeparator)
			_, statErr := os.Stat(cmdlinePath)
			if os.IsNotExist(statErr) {
				// File does not exist. That's fine. Next.
				continue
			} else if statErr != nil {
				// An error occurred. File is maybe in a Schrödinger state.
				log.Printf("Unable to stat file %s: %s", cmdlinePath, statErr)
				continue
			}

			// File exists, let's watch it
			fullPath := fmt.Sprintf("%s%c%s", outputDirectory, os.PathSeparator, f.Name())
			localFuzzers = append(localFuzzers, fullPath)
		}

		return localFuzzers
	}
}
