# AGENTS.md

Инструкции для агентов (и людей), работающих с этим репозиторием.

## Проект

`afk-fortnite` — GUI-симулятор AFK-мод для Fortnite на Go: Fyne 2.5 (окно + трей),
8 режимов нажатий, глобальные хоткеи `-` (toggle) и `=` (стоп). Один пакет
`main`, CGO обязателен (robotgo, gohook, OpenGL).

## Сборка

```bash
make build                  # текущая платформа (macOS/Windows/Linux)
make build-linux            # Linux X11
make build-linux-wayland    # Linux, нативный Wayland GUI (-tags wayland)
make build-windows          # кросс-сборка с mingw-w64
make tidy                   # go mod tidy + go mod vendor + починка vendor (см. ниже)
```

Тестов в репо нет. Проверка изменений: `gofmt -l` (кроме vendor/),
`go vet ./...` и `go build ./...` на целевых платформах. Linux-сборки удобно
гонять в Docker: `golang:1.24-bookworm`, пакеты из цели `install-deps`.

## Критичные правила

- **vendor/ коммитится в репо и нужен для NixOS** (`vendorHash = null` во
  flake). У robotgo и gohook есть C-каталоги только из заголовков (base, key,
  mouse, hook и т.д.), которые `go mod vendor` не копирует. `make tidy`
  запускает `scripts/vendor.sh`, который после vendor докладывает эти
  каталоги из GOMODCACHE. После **любого** изменения `go.mod`/`go.sum` —
  только `make tidy`, руками vendor не править.
- **GNU_SOURCE redefined** при сборке Linux — это warning от glfw
  (fyne-fork, wl_window.c), НЕ ошибка сборки. Не «чинить».
- Платформенная развилка ввода: `input.go` — общий интерфейс,
  `input_robotgo.go` (X11/Win/macOS), `input_uinput_linux.go` (uinput →
  ydotool fallback). Хоткеи: `hotkeys.go` диспетчер, `hotkeys_evdev_linux.go`
  для Wayland, gohook в `hotkeys.go` для X11. Заглушки для не-Linux — в
  `wayland_stub.go`. Новые файлы с build-тегами не ломать: `//go:build linux`
  обязателен там, где есть syscall/uinput.
- **Wayland-сборка не работает в X11-сессии** (с `-tags wayland` glfw подключается
  только к Wayland и падает при старте). Для X11 — основная сборка. Это
  задокументировано в README, менять не пытаться.
- Иконка и .desktop-файл лежат в `packaging/linux/`; на них ссылаются
  Makefile (AppImage) и `package.nix`. Путь переименовывать только вместе с
  обоими.

## Wayland (Linux)

- Симуляция ввода: uinput с правами на запись `/dev/uinput` (группа `input` или
  udev-правило, см. README), fallback — `ydotool`.
- Глобальные хоткеи: чтение `/dev/input/event*` без EVIOCGRAB; нужен рескан
  новых устройств; свои uinput-устройства (имя содержит `afk-fortnite`)
  исключаются из хоткеев, иначе фидбек-луп.
- Без прав всё должно деградировать чисто: статус-предупреждение, приложение
  живое. Пользовательские строки — на русском.

## NixOS

- `flake.nix` — пакеты `afk-fortnite` (X11), `afk-fortnite-wayland` (`-tags
  wayland`), devShell; `nixos-module.nix` — `programs.afk-fortnite`
  (enable/package/users, udev-правило на uinput + группа input).
- Флейк проверяется через `make build-nix`; сборка должна идти из in-tree
  vendor (см. правило выше).

## Коммиты

История двуязычная; пишите содержательное сообщение (по-русски или
conventional style) без шума в diff: не трогать vendor при правках кода,
не тащить личные пути и мусорные файлы (`.DS_Store`, собранные бинари — в
`.gitignore`).
