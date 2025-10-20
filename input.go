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

// Function to hold or release a key
func holdKey(key string, press bool) {
	// Check if the key exists in our map
	virtualKey, exists := keyToVirtualCode[key]
	if !exists {
		return // Key not found, do nothing
	}

	// Use robotgo.KeyToggle to press or release the key
	// KeyToggle takes the key and a modifier (we use "null" for no modifier)
	// The second parameter indicates the action: "up" to release, "down" to press
	if press {
		robotgo.KeyToggle(virtualKey, "down")
	} else {
		robotgo.KeyToggle(virtualKey, "up")
	}
}

// Function to move the mouse continuously to the right edge of the screen for a given duration
func moveMouseRightContinuously(duration time.Duration) {
	// Get screen size
	screenWidth, _ := robotgo.GetScreenSize()

	// Get current mouse position
	_, currentY := robotgo.GetMousePos()

	// Calculate the target x coordinate (right edge of the screen)
	targetX := screenWidth - 1
	// Keep the same y coordinate as current position
	targetY := currentY

	// Move the mouse smoothly to the right edge of the screen
	robotgo.MoveSmooth(targetX, targetY, 10.0, 100.0)
}
