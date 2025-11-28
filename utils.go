package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

func (app *App) generateShortID() string {
	app.dataMutex.RLock()
	defer app.dataMutex.RUnlock()

	for attempt := 0; attempt < 100; attempt++ {
		id, err := generateRandomString(4)
		if err != nil {
			continue // Or handle error appropriately
		}
		if _, exists := app.clipboardData[id]; !exists {
			return id
		}
	}

	id, _ := generateRandomString(6)
	return id
}

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

func getTempDir() string {
	return filepath.Join(os.TempDir(), "web-clipboard-go")
}

func (app *App) initTempDir() {
	err := os.MkdirAll(app.tempDir, 0755)
	if err != nil {
		panic("Failed to create temp directory: " + err.Error())
	}

	files, err := filepath.Glob(filepath.Join(app.tempDir, "*"))
	if err == nil {
		for _, file := range files {
			os.Remove(file)
		}
	}
}

// generateUUID generates a simple UUID-like string
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
