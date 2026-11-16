package main

import (
	"time"

	"github.com/go-vgo/robotgo"
)

// robotgoBackend — прежний бэкенд: X11 (XTEST) на Linux и нативный API на
// Windows/macOS. Дополнительных прав не требует.
type robotgoBackend struct{}

func (robotgoBackend) keyDown(key string) {
	// Check if the key exists in our map
	virtualKey, exists := keyToVirtualCode[key]
	if !exists {
		return // Key not found, do nothing
	}

	println("Вызов robotgo.KeyToggle() с параметрами:", virtualKey, "down")
	robotgo.KeyToggle(virtualKey, "down")
}

func (robotgoBackend) keyUp(key string) {
	// Check if the key exists in our map
	virtualKey, exists := keyToVirtualCode[key]
	if !exists {
		return // Key not found, do nothing
	}

	println("Вызов robotgo.KeyToggle() с параметрами:", virtualKey, "up")
	robotgo.KeyToggle(virtualKey, "up")
}

// Function to hold or release left mouse button
func (robotgoBackend) holdLeftMouseButton(press bool) {
	if press {
		println("Вызов robotgo.Toggle() с параметрами: left, down")
		robotgo.Toggle("left", "down")
	} else {
		println("Вызов robotgo.Toggle() с параметрами: left, up")
		robotgo.Toggle("left", "up")
	}
}

// Function to simulate a single left mouse click
func (robotgoBackend) clickLeftMouseButton() {
	println("Вызов robotgo.Click() с параметрами: left")
	robotgo.Click("left")
	time.Sleep(50 * time.Millisecond) // Small delay after click
}

// Function to move the mouse continuously to the right edge of the screen for a given duration
func (robotgoBackend) moveRightFor(duration time.Duration) {
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
