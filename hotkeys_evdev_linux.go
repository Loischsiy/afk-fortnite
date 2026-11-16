//go:build linux

package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// Глобальные хоткеи под Wayland.
//
// Wayland (в отличие от X11) не даёт приложению слушать чужие клавиши:
// расширение XRecord для XWayland композиторы отключают, а портал глобальных
// сокращений поддерживают далеко не все. Поэтому читаем события клавиатур
// напрямую из evdev — /dev/input/event*. Для этого нужны права на чтение
// этих устройств; как исправить — см. README (раздел Wayland).
//
// EVIOCGRAB не используем: устройство не захватывается монопольно, события
// по-прежнему доходят до системы, мы их только "подглядываем" — как это
// делал gohook под X11.

const (
	evTypeKey = 1 // EV_KEY

	evdevKeyMinus = 12 // KEY_MINUS — старт/стоп
	evdevKeyEqual = 13 // KEY_EQUAL — стоп

	evStateListening   = 1 // слушаем устройство
	evStateNotKeyboard = 2 // распознано, но не клавиатура
)

// startHotkeysEvdev периодически сканирует /dev/input и подключает
// слушателей ко всем найденным клавиатурам. Периодический рескан нужен,
// потому что клавиатуры могут появиться уже после старта программы
// (подключили USB, поднялся Bluetooth).
func startHotkeysEvdev() {
	var states sync.Map // /dev/input/eventX -> состояние
	firstScan := true

	scan := func() (total, listened, noPerm int) {
		events, err := filepath.Glob("/dev/input/event*")
		if err != nil {
			return 0, 0, 0
		}
		total = len(events)

		for _, dev := range events {
			if _, seen := states.Load(dev); seen {
				continue
			}

			f, err := os.Open(dev)
			if err != nil {
				// Чаще всего нет прав на чтение /dev/input/event*.
				// Молча пробуем снова при следующих сканах.
				noPerm++
				if firstScan {
					println("evdev:", dev, err.Error())
				}
				continue
			}

			name := evdevName(f)
			if !evdevIsKeyboard(f, name) {
				f.Close()
				states.Store(dev, evStateNotKeyboard)
				continue
			}

			states.Store(dev, evStateListening)
			listened++
			println("evdev: слушается", dev, "("+name+")")
			go evdevReadLoop(f, dev)
		}
		return total, listened, noPerm
	}

	total, listened, noPerm := scan()
	firstScan = false
	if total == 0 {
		hotkeysUnavailable("evdev-устройства не найдены")
	} else if listened == 0 && noPerm == total {
		hotkeysUnavailable("нет доступа на чтение к /dev/input/event* (нужна группа input, см. README)")
	}

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			scan()
		}
	}()
}

// evdevIoctl строит номер запроса _IOC(_IOC_READ, 'E', nr, len) — так в evdev
// устроены EVIOCGBIT и EVIOCGNAME.
func evdevIoctl(f *os.File, nr, length int, buf []byte) bool {
	req := uintptr(2<<30 | length<<16 | 'E'<<8 | nr)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), req, uintptr(unsafe.Pointer(&buf[0])))
	return errno == 0
}

func evdevBit(mask []byte, bit int) bool {
	return mask[bit/8]&(1<<(bit%8)) != 0
}

// evdevName возвращает имя устройства через EVIOCGNAME.
func evdevName(f *os.File) string {
	var buf [64]byte
	if !evdevIoctl(f, 0x06, len(buf), buf[:]) {
		return ""
	}
	name := string(buf[:])
	if i := strings.IndexByte(name, 0); i >= 0 {
		name = name[:i]
	}
	return strings.TrimSpace(name)
}

// evdevIsKeyboard по битовой маске EV_KEY (ioctl EVIOCGBIT, без обращения
// к sysfs) определяет, что устройство — клавиатура: должна содержать WASD,
// это отсекает мыши, тачпады и джойстики. Собственные uinput-устройства
// программы исключаем по имени, иначе её же нажатия приходили бы обратно
// как хоткеи.
func evdevIsKeyboard(f *os.File, name string) bool {
	if strings.Contains(strings.ToLower(name), "afk-fortnite") {
		return false
	}

	var types [8]byte
	if !evdevIoctl(f, 0x20, len(types), types[:]) { // EVIOCGBIT(0)
		return false
	}
	if !evdevBit(types[:], evTypeKey) {
		return false
	}

	var keys [128]byte                                      // маска кодов до 1023
	if !evdevIoctl(f, 0x20+evTypeKey, len(keys), keys[:]) { // EVIOCGBIT(EV_KEY)
		return false
	}
	for _, code := range []int{17, 30, 31, 32} { // KEY_W, KEY_A, KEY_S, KEY_D
		if !evdevBit(keys[:], code) {
			return false
		}
	}
	return true
}

// evdevReadLoop разбирает struct input_event: struct timeval (два long,
// размер зависит от разрядности архитектуры) + u16 type + u16 code + s32 value.
func evdevReadLoop(f *os.File, dev string) {
	defer f.Close()

	longSize := int(unsafe.Sizeof(uintptr(0))) // 4 на 32-бит, 8 на 64-бит
	eventSize := longSize*2 + 8

	buf := make([]byte, eventSize)
	for {
		n, err := f.Read(buf)
		if err != nil {
			println("evdev:", dev, err.Error())
			return
		}
		if n < eventSize {
			continue
		}

		evType := binary.LittleEndian.Uint16(buf[longSize*2:])
		evCode := binary.LittleEndian.Uint16(buf[longSize*2+2:])
		evValue := int32(binary.LittleEndian.Uint32(buf[longSize*2+4:]))

		// Реагируем только на само нажатие (value=1), автоповтор (2) игнорируем.
		if evType != evTypeKey || evValue != 1 {
			continue
		}

		switch evCode {
		case evdevKeyMinus:
			toggleMode()
		case evdevKeyEqual:
			stopMode()
		}
	}
}
