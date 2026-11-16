package main

import (
	"os"
	"runtime"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Бэкенды ввода
//
// robotgo на Linux отправляет ввод через X11 (XTEST), то есть работает только
// под X11: под Wayland такие события попадают максимум в XWayland и не доходят
// до нативных Wayland-приложений. Поэтому сессия определяется один раз при
// старте:
//
//   X11, Windows, macOS   — robotgo (прежнее поведение);
//   Wayland               — виртуальные устройства через /dev/uinput,
//                           а если прав на uinput нет — демон ydotool.
//
// Автодетект можно переопределить переменной окружения AFK_BACKEND=wayland|x11.
// ---------------------------------------------------------------------------

// inputBackend — низкоуровневые источники ввода, которыми пользуются режимы AFK.
type inputBackend interface {
	keyDown(key string)
	keyUp(key string)
	holdLeftMouseButton(press bool)
	clickLeftMouseButton()
	moveRightFor(duration time.Duration)
}

// activeBackend выбирается один раз в initInputBackend; по умолчанию robotgo.
var activeBackend inputBackend = robotgoBackend{}

// initInputBackend подключает нужный бэкенд до старта режимов и хоткеев.
func initInputBackend() {
	if !isWaylandSession() {
		return
	}

	b, err := newWaylandBackend()
	if err != nil {
		// Симуляция работать не будет, но GUI оставляем живым и объясняем, что делать.
		println("Wayland:", err.Error())
		statusBinding.Set("Остановлено (Wayland: нет доступа к вводу, см. README)")
		return
	}
	activeBackend = b
}

// isWaylandSession сообщает, запущена ли программа в Wayland-сессии.
// На Windows и macOS всегда false.
func isWaylandSession() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	switch os.Getenv("AFK_BACKEND") {
	case "x11":
		return false
	case "wayland":
		return true
	}
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}
	return strings.EqualFold(os.Getenv("XDG_SESSION_TYPE"), "wayland")
}

// Map to map keys to their string representations for robotgo
var keyToVirtualCode = map[string]string{
	"w":     "w",
	"a":     "a",
	"s":     "s",
	"d":     "d",
	"e":     "e",
	"space": "space",
	"shift": "shift",
}

// Обёртки, которыми пользуются режимы AFK. Логика таймингов общая для всех
// бэкендов; различается только способ доставки событий.

// Function to simulate a key press
func simulateKeyPress(key string) {
	activeBackend.keyDown(key)

	// Simulate key press with different behavior for space vs other keys
	if key == "space" {
		// Space key: short press (10ms like before)
		time.Sleep(10 * time.Millisecond)
	} else {
		// All other keys: hold for 1 second
		time.Sleep(1 * time.Second)
	}

	activeBackend.keyUp(key)
	time.Sleep(150 * time.Millisecond)
}

// Function to simulate a quick key press (for double press mode)
func quickKeyPress(key string) {
	// Quick press and release (no long hold)
	activeBackend.keyDown(key)
	time.Sleep(10 * time.Millisecond) // Very short press
	activeBackend.keyUp(key)
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
	if press {
		activeBackend.keyDown(key)
	} else {
		activeBackend.keyUp(key)
	}
}

// Function to hold or release left mouse button
func holdLeftMouseButton(press bool) {
	activeBackend.holdLeftMouseButton(press)
}

// Function to simulate a single left mouse click
func clickLeftMouseButton() {
	activeBackend.clickLeftMouseButton()
}

// Function to move the mouse continuously to the right edge of the screen for a given duration
func moveMouseRightContinuously(duration time.Duration) {
	activeBackend.moveRightFor(duration)
}
