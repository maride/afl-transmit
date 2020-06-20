package net

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"
)

var (
	peers      []Peer
	peerFile   string
	peerString string
)

// Registers flags required for peer parsing
func RegisterSenderFlags() {
	flag.StringVar(&peerFile, "peersFile", "", "File which contains the addresses for all peers, one per line")
	flag.StringVar(&peerString, "peers", "", "Addresses to peers, comma-separated.")
}

// Send the given content to all peers
func SendToPeers(content []byte) {
	for _, p := range peers {
		p.SendToPeer(content)
	}
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
