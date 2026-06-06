package main

import (
	_ "embed"
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

//go:embed icon.ico
var iconData []byte

// Global variables for script state
var (
	running       = &atomic.Bool{}
	selectedMode  = &atomic.Int32{}
	stateMu       sync.Mutex
	currentCancel context.CancelFunc
	statusBinding = binding.NewString()
)

// Human readable names shown in the UI. Index 0 maps to mode 1, etc.
var modeNames = []string{
	"AFK #1: LEGO (random WASD + space)",
	"AFK #2: AFK maps (WASD every 180s)",
	"AFK #3: Circle running (Shift+W + mouse)",
	"AFK #4: Hold S + Shift 5s + click 30s",
	"AFK #5: Hold E",
	"AFK #6: AFK maps + hold left mouse",
	"AFK #7: Shift 1s + click 30s",
	"AFK #8: Shift+W 2s then S 6s loop",
}

// modeFuncs maps the selected index (0-based) to the worker function.
var modeFuncs = []func(context.Context){
	codeOption1,
	codeOption2,
	codeOption3,
	codeOption4,
	codeOption5,
	codeOption6,
	codeOption7,
	codeOption8,
}

func main() {
	// Initialize robotgo for Windows
	robotgo.SetHandle(0)

	statusBinding.Set("Stopped")
	selectedMode.Store(0)

	a := app.NewWithID("com.afkfortnite.gui")
	icon := fyne.NewStaticResource("icon.ico", iconData)
	a.SetIcon(icon)

	w := a.NewWindow("AFK Fortnite")
	w.SetIcon(icon)

	// System tray menu (Show / Start / Stop). Fyne appends a Quit item automatically.
	if desk, ok := a.(desktop.App); ok {
		trayMenu := fyne.NewMenu("AFK Fortnite",
			fyne.NewMenuItem("Открыть окно", func() {
				w.Show()
				w.RequestFocus()
			}),
			fyne.NewMenuItem("Старт", func() { startSelectedMode() }),
			fyne.NewMenuItem("Стоп", func() { stopMode() }),
		)
		desk.SetSystemTrayMenu(trayMenu)
		desk.SetSystemTrayIcon(icon)
	}

	// Minimize to tray instead of quitting when the window is closed.
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	modeSelect := widget.NewSelect(modeNames, func(s string) {
		for i, n := range modeNames {
			if n == s {
				selectedMode.Store(int32(i))
				break
			}
		}
	})
	modeSelect.SetSelectedIndex(0)

	statusLabel := widget.NewLabelWithData(statusBinding)

	startBtn := widget.NewButton("▶  Старт", func() { startSelectedMode() })
	startBtn.Importance = widget.HighImportance
	stopBtn := widget.NewButton("■  Стоп", func() { stopMode() })

	hint := widget.NewLabel("Горячие клавиши: '-' — старт/стоп, '=' — стоп.\nЗакрытие окна сворачивает программу в системный трей.")
	hint.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewLabelWithStyle("Выберите режим AFK:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		modeSelect,
		container.NewGridWithColumns(2, startBtn, stopBtn),
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabelWithStyle("Статус:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), statusLabel),
		widget.NewSeparator(),
		hint,
	)
	w.SetContent(content)
	w.Resize(fyne.NewSize(440, 320))

	// Start global hotkey listener in the background.
	go startHotkeys()

	w.ShowAndRun()
}

// startSelectedMode stops any running mode and launches the currently selected one.
func startSelectedMode() {
	idx := int(selectedMode.Load())
	if idx < 0 || idx >= len(modeFuncs) {
		return
	}

	// Stop any previously running mode first.
	stopMode()

	ctx, cancel := context.WithCancel(context.Background())
	stateMu.Lock()
	currentCancel = cancel
	stateMu.Unlock()

	running.Store(true)
	go modeFuncs[idx](ctx)
	statusBinding.Set("Работает — " + modeNames[idx])
}

// stopMode stops the active mode and releases held keys via each worker's defer.
func stopMode() {
	running.Store(false)
	stateMu.Lock()
	if currentCancel != nil {
		currentCancel()
		currentCancel = nil
	}
	stateMu.Unlock()
	statusBinding.Set("Остановлено")
}

// toggleMode is used by the global hotkey.
func toggleMode() {
	if running.Load() {
		stopMode()
	} else {
		startSelectedMode()
	}
}

// startHotkeys keeps the global keyboard shortcuts working even when the
// window is hidden in the tray.
func startHotkeys() {
	hook.Register(hook.KeyDown, []string{"-"}, func(e hook.Event) {
		toggleMode()
	})
	hook.Register(hook.KeyDown, []string{"="}, func(e hook.Event) {
		stopMode()
	})
	s := hook.Start()
	<-hook.Process(s)
}

func codeOption1(ctx context.Context) {
	fmt.Println("AFK mode for LEGO activated")

	// Define the keys that can be randomly pressed
	keys := []string{"w", "s", "a", "d", "space"}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				// Select a random key from the available keys
				randomIndex := rand.Intn(len(keys))
				selectedKey := keys[randomIndex]

				// Press the selected key
				simulateKeyPress(selectedKey)

				// Wait for a random duration between 1-3 seconds before next key press
				randomDelay := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
				time.Sleep(randomDelay)
			} else {
				// Small pause when not running
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func codeOption2(ctx context.Context) {
	fmt.Println("AFK mode for AFK maps activated")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				doubleKeypress("w")
				doubleKeypress("s")
				doubleKeypress("a")
				doubleKeypress("d")
				select {
				case <-ctx.Done():
					return
				case <-time.After(180 * time.Second):
					// Continue after 180 seconds
				}
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func codeOption3(ctx context.Context) {
	fmt.Println("AFK mode for Circle activated")

	// Use defer to ensure keys are released when the function exits
	defer func() {
		holdKey("shift", false)
		holdKey("w", false)
	}()

	// Track the state of shift key
	shiftPressed := false

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				// If shift is not pressed yet, press and hold it along with 'w'
				if !shiftPressed {
					holdKey("shift", true)
					holdKey("w", true)
					shiftPressed = true
				}

				// Rotate mouse for 360 degrees (adjust duration as needed for smooth circle running)
				moveMouseRightContinuously(7 * time.Second)
			} else {
				// If shift is pressed and running is false, release the keys
				if shiftPressed {
					holdKey("shift", false)
					holdKey("w", false)
					shiftPressed = false
				}
				// Small pause when not running
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func codeOption4(ctx context.Context) {
	fmt.Println("AFK mode #4 activated")

	// Use defer to ensure the 's' key is released when the function exits
	defer func() {
		holdKey("s", false)
	}()

	// Create a ticker that fires every 5 seconds for Shift key
	shiftTicker := time.NewTicker(5 * time.Second)
	defer shiftTicker.Stop()

	// Create a ticker that fires every 30 seconds for mouse click
	mouseTicker := time.NewTicker(30 * time.Second)
	defer mouseTicker.Stop()

	// Start goroutine for Shift key pressing every 5 seconds
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-shiftTicker.C:
				if running.Load() {
					// Press and release 'shift' key
					holdKey("shift", true)
					time.Sleep(50 * time.Millisecond)
					holdKey("shift", false)
				}
			}
		}
	}()

	// Start goroutine for mouse clicking every 30 seconds
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-mouseTicker.C:
				if running.Load() {
					// Click left mouse button
					clickLeftMouseButton()
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				// Press and hold 's' key for 1 second
				holdKey("s", true)
				time.Sleep(3 * time.Millisecond)
				// Release 's' key
				holdKey("s", false)
				// Wait for 1 second before the next cycle
				time.Sleep(180 * time.Millisecond)
			} else {
				// Small pause when not running
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func codeOption5(ctx context.Context) {
	fmt.Println("AFK mode #5 (Hold E) activated")

	// Use defer to ensure the 'e' key is released when the function exits
	defer func() {
		holdKey("e", false)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				// Press and hold 'e' key continuously while running
				holdKey("e", true)
				// Small pause to prevent overwhelming the CPU
				time.Sleep(100 * time.Millisecond)
			} else {
				// Release 'e' key when not running
				holdKey("e", false)
			}
			// Small pause to prevent overwhelming the CPU
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Function to press a random key
func randomKeypress() {
	// Create a slice of keys to randomly select from
	keys := make([]string, 0, len(keyToVirtualCode))
	for k := range keyToVirtualCode {
		keys = append(keys, k)
	}

	// Generate random index
	randIndex := rand.Intn(len(keys))
	selectedKey := keys[randIndex]

	// Simulate key press
	simulateKeyPress(selectedKey)

	// Sleep for random duration between 1-2 seconds
	randomDelay := time.Duration(1000+rand.Intn(1000)) * time.Millisecond
	time.Sleep(randomDelay)
}

func codeOption6(ctx context.Context) {
	fmt.Println("AFK mode for AFK maps + Left Mouse Button activated")

	// Use defer to ensure left mouse button is released when the function exits
	defer func() {
		holdLeftMouseButton(false)
	}()

	// Track the state of mouse button
	mousePressed := false
	// Timer for 180 second cycles
	lastActionTime := time.Now().Add(-180 * time.Second) // Start with expired time to trigger immediate action

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				// If mouse is not pressed yet, press and hold it
				if !mousePressed {
					holdLeftMouseButton(true)
					mousePressed = true
				}

				// Check if 180 seconds have passed since last action
				if time.Since(lastActionTime) >= 180*time.Second {
					// Press keys like in mode 2
					doubleKeypress("w")
					doubleKeypress("s")
					doubleKeypress("a")
					doubleKeypress("d")
					lastActionTime = time.Now()
				}

				// Small pause to prevent overwhelming the CPU
				time.Sleep(100 * time.Millisecond)
			} else {
				// If mouse is pressed and running is false, release it
				if mousePressed {
					holdLeftMouseButton(false)
					mousePressed = false
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func codeOption7(ctx context.Context) {
	fmt.Println("AFK mode #7 (Shift every 1 seconds + Mouse click every 30 seconds) activated")

	// Create a ticker that fires every 1 seconds for Shift key
	shiftTicker := time.NewTicker(1 * time.Second)
	defer shiftTicker.Stop()

	// Create a ticker that fires every 30 seconds for mouse click
	mouseTicker := time.NewTicker(30 * time.Second)
	defer mouseTicker.Stop()

	// Start goroutine for Shift key pressing every 1 seconds
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-shiftTicker.C:
				if running.Load() {
					// Press and release 'shift' key
					simulateKeyPress("shift")
				}
			}
		}
	}()

	// Start goroutine for mouse clicking every 30 seconds
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-mouseTicker.C:
				if running.Load() {
					// Click left mouse button
					clickLeftMouseButton()
				}
			}
		}
	}()

	// Main loop just waits and keeps the function alive
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Small pause to prevent overwhelming the CPU
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func codeOption8(ctx context.Context) {
	fmt.Println("AFK mode #8 (Shift+W together, W 2s then S 6s loop) activated")

	// Ensure keys are released when exiting
	defer func() {
		holdKey("w", false)
		holdKey("s", false)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				// Step 1: Press Shift and W simultaneously; hold W for 2 seconds
				holdKey("w", true)
				// Tap shift quickly while W is held to achieve simultaneous press
				quickKeyPress("shift")
				time.Sleep(4 * time.Second)
				holdKey("w", false)

				// Step 2: Hold S for 6 seconds
				holdKey("s", true)
				time.Sleep(6 * time.Second)
				holdKey("s", false)
			} else {
				// Release any held keys when not running
				holdKey("w", false)
				holdKey("s", false)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
