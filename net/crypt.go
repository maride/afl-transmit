package net

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
)

var (
	key string
	sharedGCM cipher.AEAD
	nonceSize int
)

// RegisterCryptFlags Registers the flags required for cryptography
func RegisterCryptFlags() {
	flag.StringVar(&key, "key", "", "32 random bytes, base64-wrapped, to AES-encrypt traffic between nodes")
}

// InitCrypt creates a cipher object out of the key handed over via --key
func InitCrypt() error {
	// Check if a key was handed over
	if key == "" {
		// no key, no service
		return nil
	}

	// Unwrap base64'ed crypto bytes
	rawKey, base64Err := base64.StdEncoding.DecodeString(key)
	if base64Err != nil {
		return fmt.Errorf("failed to unpack base64'ed key: %s", base64Err)
	}

	// Create cipher object using that key
	block, cipherErr := aes.NewCipher(rawKey)
	if cipherErr != nil {
		return fmt.Errorf("failed to use your key as AES256 key: %s", cipherErr)
	}

	// Create GCM with cipher object
	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		return fmt.Errorf("failed to create GCM for your key: %s", gcmErr)
	}

	// Set shared GCM instance
	sharedGCM = gcm
	nonceSize = gcm.NonceSize()

	// No error to report
	return nil
}

// CryptApplicable checks if we are able to encrypt or decrypt things, means if this instance was given a proper key
func CryptApplicable() bool {
	// if the shared GCM object is set, we were able to get the key off user's hands, else we don't crypt
	return sharedGCM != nil
}

// Encrypt encrypts the given bytes using the symmetric key
func Encrypt(plain []byte) ([]byte, error) {
	// create nonce and fill it with random bytes
	nonce := make([]byte, nonceSize)
	_, readErr := io.ReadFull(rand.Reader, nonce)
	if readErr != nil {
		return nil, fmt.Errorf("failed to get random bytes: %s", readErr)
	}

	// Encrypt plaintext
	return sharedGCM.Seal(nonce, nonce, plain, nil), nil
}

// Decrypt decrypts the given bytes using the symmetric key
func Decrypt(enc []byte) ([]byte, error) {
	// Sanity check on input
	if len(enc) < sharedGCM.NonceSize() {
		return nil, fmt.Errorf("failed to decrypt packet: too short")
	}

	// Decrypt encrypted bytes
	return sharedGCM.Open(nil, enc[:nonceSize], enc[nonceSize:], nil)
}
