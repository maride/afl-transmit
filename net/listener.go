package net

import (
	"flag"
	"fmt"
	"github.com/maride/afl-transmit/logistic"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

var (
	port int
)

// Registers the flags required for the listener
func RegisterListenFlags() {
	flag.IntVar(&port, "port", ServerPort, "Port to bind server component to")
}

// Sets up a listener and listens forever for packets on the given port, storing their contents in the outputDirectory
func Listen(outputDirectory string) error {
	// Create listener
	addrStr := fmt.Sprintf(":%v", port)
	listener, listenErr := net.Listen("tcp", addrStr)
	if listenErr != nil {
		return listenErr
	}

	// Prepare output directory path
	outputDirectory = strings.TrimRight(outputDirectory, "/")

	// Listen forever
	for {
		// Accept connection
		conn, connErr := listener.Accept()
		if connErr != nil {
			log.Printf("Encountered error while accepting from %s: %s", conn.RemoteAddr().String(), connErr)
			continue
		}
		// Handle in a separate thread
		go handle(conn, outputDirectory)
	}
}

// Handles a single connection, and unpacks the received data into outputDirectory
func handle(conn net.Conn, outputDirectory string) {
	// Make sure to close connection on return
	defer conn.Close()

	// Read raw content
	cont, contErr := ioutil.ReadAll(conn) // bufio.NewReader(conn).ReadString('\x00')

	if contErr == nil || contErr == io.EOF {
		// We received the whole content, time to process it
		unpackErr := logistic.UnpackInto([]byte(cont), outputDirectory)
		if unpackErr != nil {
			log.Printf("Encountered error processing packet from %s: %s", conn.RemoteAddr().String(), unpackErr)
		}

		return
	} else {
		// We encountered an error on that connection
		log.Printf("Encountered error while reading from %s: %s", conn.RemoteAddr().String(), contErr)
		return
	}
}
