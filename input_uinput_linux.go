//go:build linux

package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/bendahl/uinput"
)

// evdev-коды клавиш (linux/input-event-codes.h). Те же имена, что и в
// keyToVirtualCode, но в числовом виде для uinput/ydotool.
var evdevKeyCodes = map[string]int{
	"w":     17, // KEY_W
	"e":     18, // KEY_E
	"a":     30, // KEY_A
	"s":     31, // KEY_S
	"d":     32, // KEY_D
	"shift": 42, // KEY_LEFTSHIFT
	"space": 57, // KEY_SPACE
}

// newWaylandBackend выбирает способ симуляции ввода под Wayland.
// Прямой uinput предпочтительнее; если прав на /dev/uinput нет, но запущен
// демон ydotoold — используем ydotool.
func newWaylandBackend() (inputBackend, error) {
	if b, err := newUinputBackend(); err == nil {
		println("Wayland: используется бэкенд uinput (/dev/uinput)")
		return b, nil
	} else {
		println("Wayland: uinput недоступен:", err.Error())
	}

	if b, err := newYdotoolBackend(); err == nil {
		println("Wayland: используется бэкенд ydotool")
		return b, nil
	}

	return nil, fmt.Errorf("нет доступа к /dev/uinput и не найден ydotool — как настроить, см. README (раздел Wayland)")
}

// ---------------------------------------------------------------------------
// Прямой uinput
// ---------------------------------------------------------------------------

// uinputBackend создаёт виртуальные клавиатуру и мышь в ядре через
// /dev/uinput. События неотличимы от физических устройств, поэтому работают
// одинаково и под Wayland, и под X11. Нужен доступ на запись к /dev/uinput
// (группа input, см. README).
type uinputBackend struct {
	kb    uinput.Keyboard
	mouse uinput.Mouse
}

func newUinputBackend() (*uinputBackend, error) {
	kb, err := uinput.CreateKeyboard("/dev/uinput", []byte("afk-fortnite keyboard"))
	if err != nil {
		return nil, fmt.Errorf("uinput: клавиатура: %w", err)
	}

	mouse, err := uinput.CreateMouse("/dev/uinput", []byte("afk-fortnite mouse"))
	if err != nil {
		kb.Close()
		return nil, fmt.Errorf("uinput: мышь: %w", err)
	}

	return &uinputBackend{kb: kb, mouse: mouse}, nil
}

func (b *uinputBackend) keyDown(key string) {
	code, exists := evdevKeyCodes[key]
	if !exists {
		return // Key not found, do nothing
	}

	println("Вызов uinput.Keyboard.KeyDown() с параметрами:", code)
	b.kb.KeyDown(code)
}

func (b *uinputBackend) keyUp(key string) {
	code, exists := evdevKeyCodes[key]
	if !exists {
		return // Key not found, do nothing
	}

	println("Вызов uinput.Keyboard.KeyUp() с параметрами:", code)
	b.kb.KeyUp(code)
}

// Function to hold or release left mouse button
func (b *uinputBackend) holdLeftMouseButton(press bool) {
	if press {
		println("Вызов uinput.Mouse.LeftPress()")
		b.mouse.LeftPress()
	} else {
		println("Вызов uinput.Mouse.LeftRelease()")
		b.mouse.LeftRelease()
	}
}

// Function to simulate a single left mouse click
func (b *uinputBackend) clickLeftMouseButton() {
	println("Вызов uinput.Mouse.LeftClick()")
	b.mouse.LeftClick()
	time.Sleep(50 * time.Millisecond) // Small delay after click
}

// moveRightFor непрерывно двигает курсор вправо, пока не истечёт duration.
// В режиме №3 это соответствует прежнему MoveSmooth к правому краю экрана.
func (b *uinputBackend) moveRightFor(duration time.Duration) {
	println("Вызов uinput.Mouse.MoveRight() каждые 10 мс в течение", duration)

	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		b.mouse.MoveRight(10)
		time.Sleep(10 * time.Millisecond)
	}
}

// ---------------------------------------------------------------------------
// Резервный путь: демон ydotool (ydotoold)
// ---------------------------------------------------------------------------

// ydotoolBackend — резервный вариант, когда писать в /dev/uinput напрямую
// нельзя, но установлен и запущен демон ydotoold (пакет ydotool, версия
// 1.0+). Демон работает с правами root, поэтому прав пользователю не нужно.
type ydotoolBackend struct {
	bin string
}

func newYdotoolBackend() (*ydotoolBackend, error) {
	bin, err := exec.LookPath("ydotool")
	if err != nil {
		return nil, fmt.Errorf("ydotool не найден в PATH")
	}
	return &ydotoolBackend{bin: bin}, nil
}

func (y *ydotoolBackend) run(args ...string) {
	cmd := exec.Command(y.bin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		println("ydotool:", string(out), err.Error())
	}
}

func (y *ydotoolBackend) keyDown(key string) {
	code, exists := evdevKeyCodes[key]
	if !exists {
		return // Key not found, do nothing
	}

	println("Вызов ydotool key с параметрами:", code, "down")
	y.run("key", fmt.Sprintf("%d:1", code))
}

func (y *ydotoolBackend) keyUp(key string) {
	code, exists := evdevKeyCodes[key]
	if !exists {
		return // Key not found, do nothing
	}

	println("Вызов ydotool key с параметрами:", code, "up")
	y.run("key", fmt.Sprintf("%d:0", code))
}

// Function to hold or release left mouse button
func (y *ydotoolBackend) holdLeftMouseButton(press bool) {
	if press {
		println("Вызов ydotool click с параметрами: 0x40 (down)")
		y.run("click", "0x40")
	} else {
		println("Вызов ydotool click с параметрами: 0x80 (up)")
		y.run("click", "0x80")
	}
}

// Function to simulate a single left mouse click
func (y *ydotoolBackend) clickLeftMouseButton() {
	println("Вызов ydotool click с параметрами: 0xC0 (click)")
	y.run("click", "0xC0")
	time.Sleep(50 * time.Millisecond) // Small delay after click
}

// moveRightFor двигает курсор вправо небольшими шагами, пока не истечёт duration.
func (y *ydotoolBackend) moveRightFor(duration time.Duration) {
	println("Вызов ydotool mousemove каждые 25 мс в течение", duration)

	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		y.run("mousemove", "-x", "10", "-y", "0")
		time.Sleep(25 * time.Millisecond)
	}
}
