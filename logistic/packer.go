package logistic

import (
	"archive/tar"
	"bytes"
	"compress/flate"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var noDuplicates bool

// avoidSynced is the flag to set if files containing "sync" should be packed or not.
// Those files are from another fuzzer - no need to pack them twice.
var avoidSynced bool

// RegisterPackerFlags registers flags which are required by the packer
func RegisterPackerFlags() {
	flag.BoolVar(&noDuplicates, "no-duplicates", true, "Avoid transmitting the same file multiple times, e.g. because it is present in multiple fuzzer's queues")
	flag.BoolVar(&avoidSynced, "avoid-synced", true, "Avoid transmitting files containing the keyword 'sync', as they are from other fuzzers anyways, and should be included by their afl-transmit instance")
}

// PackFuzzers packs all targeted fuzzers into a TAR - at least queue/, fuzz_bitmap, fuzzer_stats
func PackFuzzers(fuzzers []string, fuzzerDirectory string) ([]byte, error) {
	// Create TAR archive
	var tarBuffer bytes.Buffer
	tarWriter := tar.NewWriter(&tarBuffer)

	// Essentially we want to pack three things from each targeted fuzzer:
	// - the fuzz_bitmap file
	// - the fuzzer_stats file
	// - the is_main_fuzzer file if present
	// - the queue/ directory - but avoiding duplicates
	var pkgCont []string // list of queue files already present in the archive
	for _, fuzzer := range fuzzers {
		// We need full paths to read, but will write relative paths into the TAR archive
		absFuzzerPath := fuzzerDirectory
		relFuzzerPath := strings.TrimPrefix(fuzzer, fuzzerDirectory)

		// Read-n-Pack™
		packSingleFile(tarWriter, absFuzzerPath, relFuzzerPath, "fuzz_bitmap", false)
		packSingleFile(tarWriter, absFuzzerPath, relFuzzerPath, "fuzzer_stats", false)
		packSingleFile(tarWriter, absFuzzerPath, relFuzzerPath, "is_main_fuzzer", true)
		packQueueFiles(tarWriter, absFuzzerPath, relFuzzerPath, &pkgCont)
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
func packQueueFiles(tarWriter *tar.Writer, absPath string, relPath string, pkgCont *[]string) {
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

		// Check if we should care fore duplicates
		if noDuplicates && checkDuplicate(f.Name(), pkgCont) {
			// that file is already present in the package - avoid packing it again
			continue
		}

		// Check if we should care for the keyword 'sync' in file name
		if avoidSynced && strings.Contains(f.Name(), ",sync:") {
			// seems like this file was put into the queue of this fuzzer by syncing it from another fuzzer. We don't
			// need to transmit it then, because the fuzzer which found that case will have the same file but  without
			// the keyword "sync" in it. Simply put, we avoid sending the same file multiple times with different names.
			continue
		}

		// Pack into the archive
		packSingleFile(tarWriter, absPath, relPath, fmt.Sprintf("queue%c%s", os.PathSeparator, f.Name()), false)

		if noDuplicates {
			// Append added file name to the list of things included in the package
			*pkgCont = append(*pkgCont, f.Name())
		}
	}
}

// checkDuplicate checks if name is already present in pkgCont
func checkDuplicate(name string, pkgCont *[]string) bool {
	for _, p := range *pkgCont {
		if p == name {
			return true
		}
	}

	return true
}
