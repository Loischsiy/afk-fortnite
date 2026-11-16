# Аргумент self прокидывается из flake.nix.
self:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.programs.afk-fortnite;
in
{
  options.programs.afk-fortnite = {
    enable = lib.mkEnableOption "AFK Fortnite";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalMD "пакет `afk-fortnite` из этого flake (X11-сборка GUI)";
      description = ''
        Пакет для установки. По умолчанию — X11-сборка (запускается и под
        Wayland через XWayland: симуляция ввода в любом случае выбирается по
        сессии). Нативная Wayland-сборка GUI: используйте атрибут
        afk-fortnite-wayland из этого flake.
      '';
    };

    users = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      description = ''
        Пользователи, которым нужна симуляция ввода и глобальные хоткеи
        (доступ к /dev/uinput и /dev/input/event*). Под Wayland без этого
        бэкенд ввода не поднимется и хоткеи работать не будут.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    # /dev/uinput → группа uinput (правило из hardware.uinput.enable),
    # /dev/input/event* → группа input.
    hardware.uinput.enable = true;

    users.users = lib.mkIf (cfg.users != [ ]) (
      lib.genAttrs cfg.users (_: {
        extraGroups = [
          "input"
          "uinput"
        ];
      })
    );
  };
}
