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

	// Simulate key press with different behavior for space vs other keys
	robotgo.KeyDown(virtualKey)
	
	if key == "space" {
		// Space key: short press (10ms like before)
		time.Sleep(10 * time.Millisecond)
	} else {
		// All other keys: hold for 1 second
		time.Sleep(1 * time.Second)
	}
	
	robotgo.KeyUp(virtualKey)
	time.Sleep(150 * time.Millisecond)
}

// Function to simulate a quick key press (for double press mode)
func quickKeyPress(key string) {
	// Check if the key exists in our map
	virtualKey, exists := keyToVirtualCode[key]
	if !exists {
		return // Key not found, do nothing
	}

	// Quick press and release (no long hold)
	robotgo.KeyDown(virtualKey)
	time.Sleep(10 * time.Millisecond) // Very short press
	robotgo.KeyUp(virtualKey)
	time.Sleep(50 * time.Millisecond) // Short delay between actions
}

// Function to simulate a double key press
func doubleKeypress(key string) {
	// Call quickKeyPress twice with a small delay between presses
	quickKeyPress(key)
	time.Sleep(50 * time.Millisecond) // 50 milliseconds delay
	quickKeyPress(key)
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
		println("Вызов robotgo.KeyToggle() с параметрами:", virtualKey, "down")
		robotgo.KeyToggle(virtualKey, "down")
	} else {
		println("Вызов robotgo.KeyToggle() с параметрами:", virtualKey, "up")
		robotgo.KeyToggle(virtualKey, "up")
	}
}

// Function to hold or release left mouse button
func holdLeftMouseButton(press bool) {
	if press {
		println("Вызов robotgo.Toggle() с параметрами: left, down")
		robotgo.Toggle("left", "down")
	} else {
		println("Вызов robotgo.Toggle() с параметрами: left, up")
		robotgo.Toggle("left", "up")
	}
}

// Function to move the mouse continuously to the right edge of the screen for a given duration
func moveMouseRightContinuously(duration time.Duration) {
	// Get screen size
	println("Вызов robotgo.GetScreenSize()")
	screenWidth, screenHeight := robotgo.GetScreenSize()
	println("Получен размер экрана: ", screenWidth, "x", screenHeight)

	// Get current mouse position
	println("Вызов robotgo.GetMousePos()")
	currentX, currentY := robotgo.GetMousePos()
	println("Получена позиция мыши: (", currentX, ",", currentY, ")")

	// Calculate the target x coordinate (right edge of the screen)
	targetX := screenWidth - 1
	// Keep the same y coordinate as current position
	targetY := currentY

	// Move the mouse smoothly to the right edge of the screen
	println("Вызов robotgo.MoveSmooth() с параметрами:", targetX, targetY, 10.0, 100.0)
	success := robotgo.MoveSmooth(targetX, targetY, 10.0, 100.0)
	println("Результат MoveSmooth:", success)
}
