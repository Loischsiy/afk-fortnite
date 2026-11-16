# AFK Fortnite

Anti-AFK утилита с GUI (Fyne): восемь режимов имитации активности (WASD,
пробел, Shift, мышь), управление из окна или системного трея, глобальные
горячие клавиши (`-` — старт/стоп, `=` — стоп).

## Поддержка Wayland

Бэкенд симуляции ввода выбирается один раз при старте:

| Окружение                | Симуляция ввода                                    | Глобальные хоткеи              |
|--------------------------|----------------------------------------------------|--------------------------------|
| Windows, macOS           | robotgo (нативный API)                             | gohook                         |
| Linux, X11               | robotgo (XTEST)                                    | gohook (XRecord)               |
| Linux, Wayland           | uinput через `/dev/uinput`, иначе демон ydotool     | чтение `/dev/input/event*`     |

Автодетект по `WAYLAND_DISPLAY` / `XDG_SESSION_TYPE`. Можно переопределить
переменной окружения `AFK_BACKEND=wayland` или `AFK_BACKEND=x11` (пригодится
для отладки).

### Права: /dev/uinput

Основной способ под Wayland — виртуальная клавиатура и мышь через uinput.
Пользователю нужен доступ на запись к `/dev/uinput`:

```bash
sudo modprobe uinput
echo 'KERNEL=="uinput", MODE="0660", GROUP="input", OPTIONS+="static_node=uinput"' \
  | sudo tee /etc/udev/rules.d/70-afk-fortnite.rules
sudo udevadm control --reload
sudo usermod -aG input $USER   # и перелогиниться
```

Если прав на `/dev/uinput` нет, программа попробует демон `ydotoold`
(пакет `ydotool`, версия 1.0+): демон работает от root и принимает команды
через свой сокет, прав пользователю не нужно:

```bash
sudo apt install ydotool
sudo systemctl enable --now ydotool
```

### Глобальные хоткеи

Под Wayland приложение не может слушать чужие клавиши, поэтому события
клавиатур читаются напрямую из evdev (`/dev/input/event*`). Нужен доступ на
чтение — та же группа `input`, что и для uinput. Устройства не захватываются
монопольно (без EVIOCGRAB), события продолжают доходить до системы.

Без прав программа всё равно запустится: статус покажет предупреждение,
режимы работают кнопками в окне и через меню трея.

### GUI под Wayland

Стандартная Linux-сборка открывает окно через XWayland — работает на любых
композиторах, где включён XWayland. Нативная Wayland-версия GUI собирается
отдельно (`make build-linux-wayland`, Fyne собран с `-tags wayland`,
экспериментальная поддержка Fyne). В обоих случаях симуляция ввода одна и
та же — через uinput/ydotool.

Важно: wayland-сборка умеет открывать окно только при настоящей
Wayland-сессии (`WAYLAND_DISPLAY` + `XDG_RUNTIME_DIR`). В чистой X11-сессии
она при старте упадёт — для X11 используйте основную сборку.

Системный трей работает на десктопах с поддержкой StatusNotifierItem
(KDE Plasma, wlroots-композиторы с панелями типа waybar); в GNOME нужен
апплет AppIndicator.

## NixOS

Репозиторий — flake с двумя пакетами и модулем NixOS:

```bash
nix build .#afk-fortnite            # X11-версия GUI (запускается и под Wayland)
nix build .#afk-fortnite-wayland    # нативная Wayland-версия GUI
nix run .#afk-fortnite
```

Зависимости Go закоммичены в `vendor/`, сборка полностью офлайн
(`vendorHash = null`). После изменения зависимостей обязательно выполнить
`make tidy` (он запускает `scripts/vendor.sh`, который докладывает
C-заголовки robotgo/gohook, пропускаемые `go mod vendor`) и закоммитить —
flake видит только файлы, отслеживаемые git, поэтому с незакоммиченным
`vendor/` сборка не пройдёт.

Модуль NixOS устанавливает пакет, включает `hardware.uinput` и добавляет
перечисленных пользователей в группы `input` (evdev-хоткеи) и `uinput`
(`/dev/uinput`):

```nix
{
  inputs.afk-fortnite.url = "github:<ваш-логин>/afk-fortnite";

  outputs = { self, nixpkgs, afk-fortnite, ... }: {
    nixosConfigurations.my-host = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        afk-fortnite.nixosModules.default
        ({ ... }: {
          programs.afk-fortnite = {
            enable = true;
            users = [ "myuser" ];
          };
          # нативная Wayland-сборка GUI вместо X11 (опционально):
          # programs.afk-fortnite.package =
          #   afk-fortnite.packages.x86_64-linux.afk-fortnite-wayland;
        })
      ];
    };
  };
}
```

Для локальной разработки: `nix develop` даёт shell со всеми библиотеками
для CGO (X11, XTEST, GL, Wayland).

## Обычная сборка

Нужны Go 1.24+, CGO и dev-пакеты (на Ubuntu/Debian: `make install-deps`).

```bash
make build                # под текущую ОС
make build-linux          # Linux, X11-версия GUI
make build-linux-wayland  # Linux, нативная Wayland-версия GUI
make build-windows        # Windows .exe (нужен mingw-w64)
make build-gui            # Windows GUI без консоли + иконка
make build-darwin         # macOS (arm64 + amd64)
make appimage             # Linux AppImage (собирается только на Linux)
make package              # архивы для всех платформ
```

CI (`.github/workflows/build.yml`) собирает Linux (X11 + Wayland),
Windows и macOS и публикует артефакты на релиз по тегу `v*`.
