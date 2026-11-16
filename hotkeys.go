package main

import (
	hook "github.com/robotn/gohook"
)

// startHotkeys выбирает реализацию глобальных хоткеев в зависимости от
// сессии: под X11 — gohook (перехват через расширение XRecord), под Wayland —
// чтение событий клавиатур из evdev, поскольку XRecord на Wayland недоступен.
func startHotkeys() {
	if isWaylandSession() {
		startHotkeysEvdev()
		return
	}
	startHotkeysHook()
}

// startHotkeysHook keeps the global keyboard shortcuts working even when the
// window is hidden in the tray (X11, Windows, macOS).
func startHotkeysHook() {
	hook.Register(hook.KeyDown, []string{"-"}, func(e hook.Event) {
		toggleMode()
	})
	hook.Register(hook.KeyDown, []string{"="}, func(e hook.Event) {
		stopMode()
	})
	s := hook.Start()
	<-hook.Process(s)
}

// hotkeysUnavailable информирует пользователя, что горячие клавиши работать
// не будут, и как это исправить. Приложение продолжает запускаться: режимы
// можно включать кнопками и из трея.
func hotkeysUnavailable(reason string) {
	println("Глобальные хоткеи недоступны:", reason)
	statusBinding.Set("Остановлено (глобальные хоткеи недоступны: " + reason + ")")
}
