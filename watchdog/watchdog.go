package watchdog

import (
	"flag"
	"fmt"
	"github.com/maride/afl-transmit/logistic"
	"github.com/maride/afl-transmit/net"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	rescanSecs int
)

// Register flags
func RegisterWatchdogFlags() {
	flag.IntVar(&rescanSecs, "rescan-secs", 30, "Seconds to wait before rescanning local fuzzer directories")
}

// Watch over the specified directory, send updates to peers and re-scan after the specified amount of seconds
func WatchFuzzers(outputDirectory string) {
	localFuzzers := detectLocalFuzzers(outputDirectory)

	// Loop forever
	for {
		// Loop over local fuzzers
		for _, localFuzzDir := range localFuzzers {
			// Pack important parts of the fuzzer directory into a byte array
			fuzzerName := filepath.Base(localFuzzDir)
			packedFuzzer := logistic.PackFuzzer(fuzzerName, localFuzzDir)

			// and send it to our peers
			net.SendToPeers(packedFuzzer)
		}

		// Sleep a bit
		time.Sleep(time.Duration(rescanSecs) * time.Second)
	}
}

// Searches in the specified output directory for fuzzers which run locally. This is done by searching for a file which is not shared between fuzzers as it is not required for clusterized fuzzing: 'cmdline'.
func detectLocalFuzzers(outputDirectory string) []string {
	var localFuzzers []string

	// List files (read: fuzzers) in output directory
	filesInDir, readErr := ioutil.ReadDir(outputDirectory)
	if readErr != nil {
		log.Printf("Failed to list directory content of %s: %s", outputDirectory, readErr)
		return nil
	}

	// Walk over each and search for 'cmdline' file
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
