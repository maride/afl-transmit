package logistic

import (
	"archive/tar"
	"bytes"
	"compress/flate"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// PackFuzzers packs all targeted fuzzers into a TAR - at least queue/, fuzz_bitmap, fuzzer_stats
func PackFuzzers(fuzzers []string, fuzzerDirectory string) ([]byte, error) {
	// Create TAR archive
	var tarBuffer bytes.Buffer
	tarWriter := tar.NewWriter(&tarBuffer)

	// Essentially we want to pack three things from each targeted fuzzer:
	// - the fuzz_bitmap file
	// - the fuzzer_stats file
	// - the is_main_fuzzer file if present
	// - the queue/ directory
	for _, fuzzer := range fuzzers {
		// We need full paths to read, but will write relative paths into the TAR archive
		absFuzzerPath := fuzzerDirectory
		relFuzzerPath := strings.TrimPrefix(fuzzer, fuzzerDirectory)

		// Read-n-Pack™
		packSingleFile(tarWriter, absFuzzerPath, relFuzzerPath, "fuzz_bitmap", false)
		packSingleFile(tarWriter, absFuzzerPath, relFuzzerPath, "fuzzer_stats", false)
		packSingleFile(tarWriter, absFuzzerPath, relFuzzerPath, "is_main_fuzzer", true)
		packQueueFiles(tarWriter, absFuzzerPath, relFuzzerPath)
	}

	// Close TAR archive
	tarWriter.Close()

	// Prepare FLATE compression
	var flateBuffer bytes.Buffer
	flateWrite, flateErr := flate.NewWriter(&flateBuffer, flate.BestCompression)
	if flateErr != nil {
		return nil, fmt.Errorf("unable to prepare flate compressor: %s", flateErr)
	}

	// Apply FLATE compression
	flateWrite.Write(tarBuffer.Bytes())
	flateWrite.Close()

	// Return result: a DEFLATEd TAR archive
	return flateBuffer.Bytes(), nil
}

// packSingleFile packs a single file and writes it to the archive
// fuzzerDirectory is the base directory, e.g. /project/fuzzers/
// fuzzer is the name of the fuzzer itself, e.g. main-fuzzer-01
// filename is the name of the file you want to pack, e.g. fuzzer_stats
// ignoreNotFound is just used for files which may not be present in all fuzzer directories, like is_main_fuzzer
func packSingleFile(tarWriter *tar.Writer, absPath string, relPath string, fileName string, ignoreNotFound bool) {
	// Read file
	readPath := fmt.Sprintf("%s%c%s%c%s", absPath, os.PathSeparator, relPath, os.PathSeparator, fileName)
	contents, readErr := ioutil.ReadFile(readPath)
	if readErr != nil {
		if !ignoreNotFound {
			log.Printf("Failed to read file %s: %s", readPath, readErr)
		}
		return
	}

	// Create header for this file
	header := &tar.Header{
		Name: fmt.Sprintf("%s%c%s", relPath, os.PathSeparator, fileName),
		Mode: 0600,
		Size: int64(len(contents)),
	}

	// Add header and contents to archive
	tarWriter.WriteHeader(header)
	tarWriter.Write(contents)
}

// Packs the files in the given directory into a tar archive
func packQueueFiles(tarWriter *tar.Writer, absPath string, relPath string) {
	// Get list of queue files
	queuePath := fmt.Sprintf("%s%c%s%cqueue", absPath, os.PathSeparator, relPath, os.PathSeparator)
	filesInDir, readErr := ioutil.ReadDir(queuePath)
	if readErr != nil {
		log.Printf("Failed to list directory content of %s: %s", queuePath, readErr)
		return
	}

	// Walk over each file and add it to our archive
	for _, f := range filesInDir {
		// Check if we hit a directory (e.g. '.state')
		if f.IsDir() {
			// Ignore directories altogether
			continue
		}

		// Pack into the archive
		packSingleFile(tarWriter, absPath, relPath, fmt.Sprintf("queue%c%s", os.PathSeparator, f.Name()), false)
	}
}
