package main

import (
	"flag"
	"fmt"
	"github.com/maride/afl-transmit/logistic"
	"github.com/maride/afl-transmit/net"
	"github.com/maride/afl-transmit/stats"
	"github.com/maride/afl-transmit/watchdog"
	"log"
)

var (
	outputDirectory string
	printStats bool
)

func main() {
	// Register flags
	watchdog.RegisterWatchdogFlags()
	net.RegisterSenderFlags()
	net.RegisterListenFlags()
	net.RegisterCryptFlags()
	logistic.RegisterPackerFlags()
	RegisterGlobalFlags()
	flag.Parse()

	// Check if we have the only required argument present - outputDirectory
	if outputDirectory == "" {
		fmt.Println("Please specify fuzzer-directory. See help (--help) for details.")
		return
	}

	// Read peers file
	net.ReadPeers()

	// Initialize crypto if desired
	cryptErr := net.InitCrypt()
	if cryptErr != nil {
		fmt.Printf("Failed to initialize crypt function: %s", cryptErr)
		return
	}

	// Start watchdog for local afl instances
	go watchdog.WatchFuzzers(outputDirectory)

	// Start stat printer
	if printStats {
		go stats.PrintStats()
	}

	// Listen for incoming connections
	listenErr := net.Listen(outputDirectory)
	if listenErr != nil {
		log.Println(listenErr)
	}
}

// Registers flags which are required by multiple modules and need to be handled here
func RegisterGlobalFlags() {
	flag.StringVar(&outputDirectory, "fuzzer-directory", "", "The output directory of the fuzzer(s)")
	flag.BoolVar(&printStats, "print-stats", true, "Print traffic statistics every few seconds")
}
