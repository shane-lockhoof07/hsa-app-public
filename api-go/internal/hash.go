package internal

import (
	"crypto/sha256"
	"fmt"
)

// HashImage generates a SHA256 hash of the image data
func HashImage(imageData []byte) string {
	hash := sha256.Sum256(imageData)
	return fmt.Sprintf("%x", hash)
}