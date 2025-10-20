package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

// Global variables for script state
var (
	running       = &atomic.Bool{}
	stopRequested = &atomic.Bool{}
)

func main() {
	// Initialize robotgo for Windows
	robotgo.SetHandle(0)
	mainMenu()
}

func mainMenu() {
	fmt.Println("*******************************")
	fmt.Println("*                             *")
	fmt.Println("*     SELECT AFK OPTION       *")
	fmt.Println("*                             *")
	fmt.Println("*******************************")
	fmt.Println("Choose which AFK mode you want to use:")
	fmt.Println("[1] - AFK #1: AFK mode for LEGO")
	fmt.Println("[2] - AFK #2: AFK mode for AFK maps")
	fmt.Println("[3] - AFK #3: AFK mode for Circle runing")
	fmt.Println("[4] - AFK #4: S-Press + Shift")
	fmt.Println("[5] - AFK #5: Hold E")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	choice, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid input. Please enter a number.")
		return
	}

	switch choice {
	case 1:
		ctx, cancel := context.WithCancel(context.Background())
		// Start key listener in a separate goroutine
		go keyListener(cancel)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			codeOption1(ctx)
		}()
		wg.Wait()
		cancel()
	case 2:
		ctx, cancel := context.WithCancel(context.Background())
		// Start key listener in a separate goroutine
		go keyListener(cancel)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			codeOption2(ctx)
		}()
		wg.Wait()
		cancel()
	case 3:
		ctx, cancel := context.WithCancel(context.Background())
		// Start key listener in a separate goroutine
		go keyListener(cancel)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			codeOption3(ctx)
		}()
		wg.Wait()
		cancel()
	case 4:
		ctx, cancel := context.WithCancel(context.Background())
		// Start key listener in a separate goroutine
		go keyListener(cancel)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			codeOption4(ctx)
		}()
		wg.Wait()
		cancel()
	case 5:
		ctx, cancel := context.WithCancel(context.Background())
		// Start key listener in a separate goroutine
		go keyListener(cancel)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			codeOption5(ctx)
		}()
		wg.Wait()
		cancel()
	default:
		fmt.Println("Invalid choice. Please try again.")
		return
	}
}

func codeOption1(ctx context.Context) {
	fmt.Println("AFK mode for LEGO activated")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if running.Load() {
				randomKeypress()
			}
			time.Sleep(100 * time.Millisecond)
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

	// Create a ticker that fires every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if running.Load() {
				// Press and release 'shift' key
				holdKey("shift", true)
				time.Sleep(50 * time.Millisecond)
				holdKey("shift", false)
			}
		default:
			if running.Load() {
				// Press and hold 's' key for 0.3 seconds
				holdKey("s", true)
				time.Sleep(300 * time.Millisecond)
				// Release 's' key
				holdKey("s", false)
				// Sleep for 0.3 seconds before the next cycle
				time.Sleep(300 * time.Millisecond)
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

// Toggle the script on/off
func toggleScript() {
	newState := !running.Load()
	running.Store(newState)
	if newState {
		fmt.Println("Script Started")
	} else {
		fmt.Println("Script Stopped")
	}
}

// Stop the script completely
func stopScript(cancel context.CancelFunc) {
	running.Store(false)
	stopRequested.Store(true)
	fmt.Println("Script Stopped")
	cancel()
}

// Key listener function
func keyListener(cancel context.CancelFunc) {
	// Register key event for '=' key to toggle script
	hook.Register(hook.KeyDown, []string{"="}, func(e hook.Event) {
		toggleScript()
	})

	// Register key event for '-' key to stop script
	hook.Register(hook.KeyDown, []string{"-"}, func(e hook.Event) {
		stopScript(cancel)
		hook.End()
	})

	// Start the hook for processing events
	s := hook.Start()
	<-hook.Process(s)
}
