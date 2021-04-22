package logistic

import (
	"archive/tar"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
)

// UnpackInto decrompesses the given bytes with DEFLATE, then unpacks the result as TAR archive into the targetDir
func UnpackInto(raw []byte, targetDir string) error {
	// Prepare FLATE decompressor
	var flateBuffer bytes.Buffer
	flateReader := flate.NewReader(&flateBuffer)

	// Uncompress
	flateBuffer.Write(raw)
	raw, _ = ioutil.ReadAll(flateReader)

	// Open TAR archive
	var tarBuffer bytes.Buffer
	tarBuffer.Write(raw)
	tarReader := tar.NewReader(&tarBuffer)

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

	return nil
}

// Writes the contents to the target
func unpackSingleFile(raw []byte, targetDirectory string, filename string) {
	destPath := fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, filename)

	// Check if the file already exists - we won't overwrite it then
	_, fileInfoErr := os.Stat(destPath)
	if os.IsExist(fileInfoErr) {
		// File already exists, we don't need to write a thing
		return
	}

	// Check if the target directory already exists - otherwise we create it
	dirOfFile := path.Dir(fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, filename))
	_, dirInfoErr := os.Stat(dirOfFile)
	if os.IsNotExist(dirInfoErr) {
		// Create directories as required
		mkdirErr := os.MkdirAll(dirOfFile, 0755)
		if mkdirErr != nil {
			log.Printf("Failed to create directory %s: %s", dirOfFile, mkdirErr)
			return
		}
	}

	// Write file
	writeErr := ioutil.WriteFile(destPath, raw, 0644)
	if writeErr != nil {
		log.Printf("Unable to write to file %s: %s", destPath, writeErr)
	}
}
