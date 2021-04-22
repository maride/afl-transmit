package logistic

import (
	"archive/tar"
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Unpacks a raw string, creates files and stores them in the target directory. May return an error if one occurrs
func UnpackInto(raw []byte, targetDir string) error {
	// Prepare FLATE decompressor
	var flateBuffer bytes.Buffer
	flateReader := flate.NewReader(&flateBuffer)

	// Uncompress
	flateBuffer.Write(raw)
	raw, _ = ioutil.ReadAll(flateReader)

	// Process raw bytes
	splitted := bytes.Split(raw, []byte("\n"))
	if len(splitted) != 4 {
		// We are currently packing four things in there (the fuzzer name, queue/, fuzz_bitmap, fuzzer_stats)
		// So if we don't get three parts, we have a malformed packet
		return fmt.Errorf("unable to unpack packet: Expected 4 parts, got %d", len(splitted))
	}

	// base64 decode contents
	for i, s := range splitted {
		b64Buf := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
		base64.StdEncoding.Decode(b64Buf, s)
		splitted[i] = b64Buf
	}

	// Check filename, and process it
	fuzzerName := string(bytes.TrimRight(splitted[0], "\x00"))
	if strings.Contains(fuzzerName, "/") {
		return fmt.Errorf("received file name with a slash, discarding whole packet for fuzzer \"%s\"", fuzzerName)
	}

	// Check if our target directory (this very fuzzers directory) already exists, or if we need to create it
	targetDir = fmt.Sprintf("%s%c%s", targetDir, os.PathSeparator, fuzzerName)
	_, folderErr := os.Stat(targetDir)
	if os.IsNotExist(folderErr) {
		// directory doesn't yet exist, create it
		mkdirErr := os.MkdirAll(targetDir, 0700)
		if mkdirErr != nil {
			// Creating the target directory failed, so we won't proceed unpacking into a non-existent directory
			return fmt.Errorf("unable to unpack packet: could not create directory at %s: %s", targetDir, mkdirErr)
		}
	}

	// Process every single part
	unpackSingleFile(splitted[1], targetDir, "fuzz_bitmap")
	unpackSingleFile(splitted[2], targetDir, "fuzzer_stats")
	unpackQueueDir(splitted[3], targetDir)

	return nil
}

// Writes the contents to the target
func unpackSingleFile(raw []byte, targetDirectory string, filename string) {
	path := fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, filename)

	// Check if the file already exists - we won't overwrite it then
	_, fileInfoErr := os.Stat(path)
	if os.IsExist(fileInfoErr) {
		// File already exists, we don't need to write a thing
		return
	}

	writeErr := ioutil.WriteFile(path, raw, 0644)
	if writeErr != nil {
		log.Printf("Unable to write to file %s: %s", path, writeErr)
	}
}

// Writes all files in the raw byte array into the target directory
func unpackQueueDir(raw []byte, targetDir string) {
	// Open TAR archive
	var tarBuffer bytes.Buffer
	tarBuffer.Write(raw)
	tarReader := tar.NewReader(&tarBuffer)

	// Set correct path for files
	targetDir = fmt.Sprintf("%s%cqueue", targetDir, os.PathSeparator)

	// Create queue directory if it doesn't exist yet
	_, folderErr := os.Stat(targetDir)
	if os.IsNotExist(folderErr) {
		os.Mkdir(targetDir, 0755)
	}

	// Iterate over all files in the archive
	for {
		// Read header
		header, headerErr := tarReader.Next()
		if headerErr == io.EOF {
			// We reached the end of the TAR archive. Fine.
			break
		} else if headerErr != nil {
			// Unknown error occurred
			log.Printf("Error parsing TAR header entry: %s", headerErr)
			break
		}

		// Write file
		var fileBuffer bytes.Buffer
		io.Copy(&fileBuffer, tarReader)
		unpackSingleFile(fileBuffer.Bytes(), targetDir, header.Name)
	}
}
