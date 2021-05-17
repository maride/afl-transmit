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
)

// RegisterWatchdogFlags registers required flags for the watchdog
func RegisterWatchdogFlags() {
	flag.IntVar(&rescan, "rescan", 30, "Minutes to wait before rescanning local fuzzer directory")
}

// WatchFuzzers watches over the specified directory, sends updates to peers and re-scans after the specified amount of seconds
func WatchFuzzers(outputDirectory string) {
	// Loop forever
	for {
		// Search for main fuzzer
		targetFuzzer, targetErr := getTargetFuzzer(outputDirectory)
		if targetErr != nil {
			log.Printf("Failed to detect main fuzzer: %s", targetErr)
			continue
		}

		// Pack important parts of the fuzzer into an archive
		packedFuzzers, packerErr := logistic.PackFuzzer(targetFuzzer, outputDirectory)
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

// Searches in the specified output directory for the main fuzzer.
// Identifying the main fuzzer is done by searching for the file "is_main_node". On secondary-only servers, this relies
// on the "election process" done by secondary fuzzers if they don't find a local main node. In that election process, a
// secondary fuzzer becomes the main node in terms of syncing. That means we need to focus on the main node even if
// there is no real main node running locally. Good thing is that the elected node will create a "is_main_node" file, so
// we really just need to search for that.
// However it is important to avoid transmitting the is_main_node file - else, the secondaries will think that there is
// a main node running locally. Skipping the is_main_node file means the other fuzzers (and especially the elected main
// fuzzer) detect the transmitted fuzzer as a normal (and dead) secondary.
//
// (see AFLplusplus/src/afl-fuzz.run.c:542)
func getTargetFuzzer(outputDirectory string) (string, error) {
	// find main fuzzer directory - its the only one required by the secondaries

	// List files (read: fuzzers) in output directory
	filesInDir, readErr := ioutil.ReadDir(outputDirectory)
	if readErr != nil {
		return "", fmt.Errorf("Failed to list directory content of %s: %s", outputDirectory, readErr)
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
		return fmt.Sprintf("%s%c%s", outputDirectory, os.PathSeparator, f.Name()), nil
	}

	// Failed to find the main node - probably we are in --main mode by accident
	return "", fmt.Errorf("Unable to find main node in %s", outputDirectory)
}
