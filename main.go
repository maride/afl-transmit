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
)

func main() {
	// Register flags
	watchdog.RegisterWatchdogFlags()
	net.RegisterSenderFlags()
	net.RegisterListenFlags()
	net.RegisterCryptFlags()
	logistic.RegisterPackerFlags()
	stats.RegisterStatsFlags()
	RegisterGlobalFlags()
	flag.Parse()

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
	go stats.PrintStats()

	// Listen for incoming connections
	listenErr := net.Listen(outputDirectory)
	if listenErr != nil {
		log.Println(listenErr)
	}
}

// Registers flags which are required by multiple modules and need to be handled here
func RegisterGlobalFlags() {
	flag.StringVar(&outputDirectory, "fuzzer-directory", "", "The output directory of the fuzzer(s)")
}
