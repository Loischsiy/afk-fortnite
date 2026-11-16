{
  lib,
  buildGoModule,
  pkg-config,
  copyDesktopItems,
  makeDesktopItem,
  libGL,
  libpng,
  wayland,
  libxkbcommon,
  wayland-protocols,
  xorg,
  version,
  src,
  withWayland ? false,
}:

# Сборка для NixOS. Два варианта:
#   afk-fortnite          — X11-версия GUI (работает и под Wayland через
#                           XWayland; симуляция ввода на Wayland всё равно
#                           идёт через uinput/ydotool);
#   afk-fortnite-wayland  — нативная Wayland-версия GUI (Fyne собран с
#                           -tags wayland; экспериментально).
# Зависимости лежат в коммите репозитория (vendor/), поэтому сборка полностью
# офлайн и vendorHash не нужен.

buildGoModule {
  pname = "afk-fortnite" + lib.optionalString withWayland "-wayland";
  inherit version src;

  vendorHash = null; # используется vendor/ из исходников

  env.CGO_ENABLED = "1";

  nativeBuildInputs = [
    pkg-config
    copyDesktopItems
  ];

  buildInputs =
    [
      libGL
      libpng
      xorg.libX11
      xorg.libXtst
      xorg.libXrandr
      xorg.libXinerama
      xorg.libXcursor
      xorg.libXfixes
      xorg.libXi
      xorg.libXext
      xorg.libXrender
      xorg.xorgproto
    ]
    ++ lib.optionals withWayland [
      wayland
      libxkbcommon
      wayland-protocols
    ];

  tags = lib.optionals withWayland [ "wayland" ];

  ldflags = [
    "-s"
    "-w"
  ];

  # Иконка и .desktop для меню приложений.
  postInstall = ''
    install -Dm644 packaging/linux/afk-fortnite.png \
      $out/share/icons/hicolor/256x256/apps/afk-fortnite.png
  '';

  desktopItems = [
    (makeDesktopItem {
      name = "afk-fortnite";
      exec = "afk-fortnite";
      icon = "afk-fortnite";
      desktopName = "AFK Fortnite";
      comment = "Anti-AFK assistant for Fortnite";
      categories = [ "Game" ];
    })
  ];

  meta = with lib; {
    description = "Anti-AFK assistant for Fortnite (GUI)";
    license = licenses.gpl3Only;
    mainProgram = "afk-fortnite";
    platforms = platforms.linux;
  };
}
