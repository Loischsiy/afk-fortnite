package main

import (
	"time"

	"github.com/go-vgo/robotgo"
)

// Map to map keys to their string representations for robotgo
var keyToVirtualCode = map[string]string{
	"w":     "w",
	"a":     "a",
	"s":     "s",
	"d":     "d",
	"space": "space",
}

// Function to simulate a key press
func simulateKeyPress(key string) {
	// Check if the key exists in our map
	virtualKey, exists := keyToVirtualCode[key]
	if !exists {
		return // Key not found, do nothing
	}

	// Simulate key press and release using robotgo
	robotgo.KeyTap(virtualKey)
}

// Function to simulate a double key press
func doubleKeypress(key string) {
	// Call simulateKeyPress twice with a small delay between presses
	simulateKeyPress(key)
	time.Sleep(50 * time.Millisecond) // 50 milliseconds delay
	simulateKeyPress(key)
}
