package net

import (
	"flag"
	"github.com/maride/afl-transmit/stats"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"
)

var (
	peers      []Peer
	peerFile   string
	peerString string
	removeLocals bool
)

// Registers flags required for peer parsing
func RegisterSenderFlags() {
	flag.StringVar(&peerFile, "peersFile", "", "File which contains the addresses for all peers, one per line")
	flag.StringVar(&peerString, "peers", "", "Addresses to peers, comma-separated.")
	flag.BoolVar(&removeLocals, "remove-locals", false, "Skip addresses which are served on local interfaces. This allows you to use the same peer file for all of your hosts. Please note that not too much effort is spent on resolving conflicts. If you are e.g. giving hostnames as peers, filtering won't work as expected.")
}

// Send the given content to all peers
func SendToPeers(content []byte) {
	// Reset stats
	alivePeers := uint8(0)

	// Peers where the sending process initially failed
	var failedPeers []Peer

	for i := 0; i < len(peers); i++ {
		// Send to that peer
		sendErr := peers[i].SendToPeer(content)
		if sendErr != nil {
			// Sending failed, retry in a second
			failedPeers = append(failedPeers, peers[i])
			continue
		}
		alivePeers++
	}

	// Set stats now - we might be updating it in a few minutes again
	stats.SetAlivePeers(alivePeers)

	// Sleep so our peers maybe come up again
	if len(failedPeers) > 0 {
		time.Sleep(10 * time.Second)
	}

	// Retry
	for i := 0; i < len(failedPeers); i++ {
		// Send to that peer
		sendErr := failedPeers[i].SendToPeer(content)
		if sendErr != nil {
			// Sending failed - inform user
			log.Printf("Transmission failed after retry: %s", sendErr)
			continue
		}
		alivePeers++
	}

	// Finally update it again, now including those peers where we retried it
	stats.SetAlivePeers(alivePeers)
}

// Parses both peerString and peerFile, and adds all the peers to an internal array.
func ReadPeers() {
	// Read peer file if it is given
	if peerFile != "" {
		fileErr := readPeersFile(peerFile)

		// Check if we encountered errors
		if fileErr != nil {
			log.Printf("Failed to read peer file: %s", fileErr)
		}
	}

	// Read peer string if it is given
	if peerString != "" {
		readPeersString(peerString)
	}

	// Remove doubles.
	removeDoubledPeers()

	// Remove locally bound if requested
	if removeLocals {
		removeLocalPeers()
	}

	// Update stats, include registered peers
	stats.PushStat(stats.Stat{
		RegisteredPeers: uint8(len(peers)),
	})

	log.Printf("Configured %d unique peers.", len(peers))
}

// Read a peer file at the given path, parses it and adds newly created Peers to the internal peers array
func readPeersFile(path string) error {
	// Read file
	readContBytes, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return readErr
	}

	// Convert to string
	readCont := string(readContBytes)

	// Iterate over it, line by line
	for _, line := range strings.Split(readCont, "\n") {
		// Check if line is usable
		if len(line) == 0 {
			// Empty line, ignore
			continue
		}
		// Append newly created peer to array
		peers = append(peers, CreatePeer(line))
	}

	return nil
}

// Read peers from the given string, parses it and adds newly created Peers to the internal peers array
func readPeersString(raw string) {
	for _, peer := range strings.Split(raw, ",") {
		// Append newly created peer to array
		peers = append(peers, CreatePeer(peer))
	}
}

// Iterates over the peers array and removes doubles
func removeDoubledPeers() {
	// Outer loop - go over all peers
	for i := 0; i < len(peers); i++ {
		// Inner loop - go over peers after the current (i) one, removing those with the same address
		for j := i + 1; j < len(peers); j++ {
			if peers[j].Address == peers[i].Address {
				// Double found, remove j'th element
				peers = append(peers[:j], peers[j+1:]...)
			}
		}
	}
}

// Removes local peers, means addresses which are present on local interfaces, from the peers array.
func removeLocalPeers() {
	interfaces, interfacesErr := net.Interfaces()
	if interfacesErr != nil {
		log.Printf("Unable to remove local peers because interface lookup failed: %s", interfacesErr)
		return
	}

	// Iterate over all interfaces, and collect all addresses
	var blacklistPeers []Peer
	for _, i := range interfaces {
		// Get all addresses of this interface
		iAddrs, addrsErr := i.Addrs()
		if addrsErr != nil {
			log.Printf("Unable to get address of interface %s: %s", i.Name, addrsErr)
			continue
		}

		// Append all addresses
		for _, a := range iAddrs {
			blPeer := CreatePeer(a.(*net.IPNet).IP.String())
			blacklistPeers = append(blacklistPeers, blPeer)
		}
	}

	// Check all peers against all addresses
	for i := 0; i < len(peers); i++ {
		for _, p := range blacklistPeers {
			if peers[i].Address == p.Address {
				// Found match, remove that peer
				peers = append(peers[:i], peers[i+1:]...)
				i--
				break
			}
		}
	}
}
