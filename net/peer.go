package net

import (
	"log"
	"net"
)

type Peer struct {
	Address string
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
	_, writeErr := tcpConn.Write(content)
	if writeErr != nil {
		log.Printf("Unable to write to peer %s: %s", tcpConn.RemoteAddr().String(), writeErr)
		return
	}

	// Close connection
	tcpConn.Close()
}
