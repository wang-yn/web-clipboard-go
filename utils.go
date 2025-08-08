package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (app *App) generateShortID() string {
	rand.Seed(time.Now().UnixNano())

	app.dataMutex.RLock()
	defer app.dataMutex.RUnlock()

	for attempt := 0; attempt < 100; attempt++ {
		id := make([]byte, 4)
		for i := range id {
			id[i] = chars[rand.Intn(len(chars))]
		}
		idStr := string(id)

		if _, exists := app.clipboardData[idStr]; !exists {
			return idStr
		}
	}

	id := make([]byte, 6)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return string(id)
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