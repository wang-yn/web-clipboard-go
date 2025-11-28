package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateShortID generates a random short ID for clipboard items
func GenerateShortID(existingIDs map[string]bool) string {
	for attempt := 0; attempt < 100; attempt++ {
		id, err := generateRandomString(4)
		if err != nil {
			continue
		}
		if !existingIDs[id] {
			return id
		}
	}

	id, _ := generateRandomString(6)
	return id
}

// GenerateRandomString generates a random alphanumeric string of the given length
func generateRandomString(length int) (string, error) {
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(chars[num.Int64()])
	}
	return sb.String(), nil
}

// GenerateUUID generates a simple UUID-like string
func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
