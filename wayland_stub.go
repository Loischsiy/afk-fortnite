//go:build !linux

package main

import "errors"

// Заглушки для платформ, где Wayland не существует (Windows, macOS): символы
// нужны компилятору, но по факту никогда не вызываются — isWaylandSession()
// на этих системах всегда false.

func newWaylandBackend() (inputBackend, error) {
	return nil, errors.New("Wayland поддерживается только на Linux")
}

func startHotkeysEvdev() {
	hotkeysUnavailable("Wayland поддерживается только на Linux")
}
