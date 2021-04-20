package net

import (
	"fmt"
	"github.com/maride/afl-transmit/stats"
	"log"
	"net"
	"regexp"
	"strings"
)

var (
	portSuffixRegex = regexp.MustCompile(":\\d{0,5}$")
)

type Peer struct {
	Address string
}

// Creates a peer from the given address
func CreatePeer(address string) Peer {
	// Clean line
	address = strings.TrimSpace(address)

	// Check if a port is already part of the address
	// This is the lazy way: if a IPv6 literal is given without square brackets and without a port, this will fail badly.
	if !portSuffixRegex.MatchString(address) {
		// Port number is not yet part of the address, so append the default port number
		address = fmt.Sprintf("%s:%d", address, ServerPort)
	}

	// Return constructed Peer
	return Peer{
		Address: address,
	}
}

// Sends the given content to the peer
func (p *Peer) SendToPeer(content []byte) {
	// Build up a connection
	tcpConn, dialErr := net.Dial("tcp", p.Address)
	if dialErr != nil {
		log.Printf("Unable to connect to peer %s: %s", p.Address, dialErr)
		return
	}

	// Send
	written, writeErr := tcpConn.Write(content)
	if writeErr != nil {
		log.Printf("Unable to write to peer %s: %s", tcpConn.RemoteAddr().String(), writeErr)
		return
	}

	// Push written bytes to stats
	stats.PushStat(stats.Stat{SentBytes: uint64(written)})

	// Close connection
	tcpConn.Close()
}
