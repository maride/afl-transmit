package logistic

import (
	"archive/tar"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Packs a whole fuzzer directory - at least queue/, fuzz_bitmap, fuzzer_stats
func PackFuzzer(fuzzerName string, directory string) []byte {
	// Gather contents
	contentArray := [][]byte{
		[]byte(fuzzerName),
		packSingleFile(directory, "fuzz_bitmap"),
		packSingleFile(directory, "fuzzer_stats"),
		packQueueFiles(directory),
	}

	// Convert all parts to base64, and concat them to the packet
	var result []byte
	for _, a := range contentArray {
		b64Buf := make([]byte, base64.StdEncoding.EncodedLen(len(a)))
		base64.StdEncoding.Encode(b64Buf, a)

		// Add newline char as separator
		result = append(result, '\n')

		// Append base64 encoded content
		result = append(result, b64Buf...)
	}

	// Return result: a big byte array, representing concatted base64-encoded files
	return result
}

// Reads a single file and returns it
func packSingleFile(directory string, fileName string) []byte {
	path := fmt.Sprintf("%s%c%s", directory, os.PathSeparator, fileName)
	contents, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		log.Printf("Failed to read file %s: %s", path, readErr)
		return nil
	}

	return contents
}

// Packs the files in the given directory into a tar archive
func packQueueFiles(directory string) []byte {
	var tarBuffer bytes.Buffer
	tarWriter := tar.NewWriter(&tarBuffer)

	// Get list of queue files
	queuePath := fmt.Sprintf("%s%cqueue", directory, os.PathSeparator)
	filesInDir, readErr := ioutil.ReadDir(queuePath)
	if readErr != nil {
		log.Printf("Failed to list directory content of %s: %s", directory, readErr)
		return nil
	}

	// Walk over each file and add it to our archive
	for _, f := range filesInDir {
		// Check if we hit a directory (e.g. '.state')
		if f.IsDir() {
			// Ignore directories altogether
			continue
		}

		// Create header for this file
		header := &tar.Header{
			Name: f.Name(),
			Mode: 0600,
			Size: f.Size(),
		}

		// Read file
		path := fmt.Sprintf("%s%c%s", queuePath, os.PathSeparator, f.Name())
		contents, readErr := ioutil.ReadFile(path)
		if readErr != nil {
			log.Printf("Failed to read file %s: %s", path, readErr)
			continue
		}

		// Add header and contents to archive
		tarWriter.WriteHeader(header)
		tarWriter.Write(contents)
	}

	// Close constructed tar archive
	tarWriter.Close()

	// And return it
	return tarBuffer.Bytes()
}
