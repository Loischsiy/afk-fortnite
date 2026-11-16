{
  description = "AFK Fortnite — anti-AFK утилита (GUI на Fyne, симуляция ввода для X11 и Wayland)";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
      version = self.shortRev or "dev";
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        rec {
          afk-fortnite = pkgs.callPackage ./package.nix {
            inherit version;
            src = self;
          };

          # Нативная Wayland-сборка GUI (Fyne с -tags wayland).
          afk-fortnite-wayland = afk-fortnite.override { withWayland = true; };

          default = afk-fortnite;
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages =
              with pkgs;
              [
                go
                gcc
                pkg-config
              ]
              ++ (with xorg; [
                libX11
                libXtst
                libXrandr
                libXinerama
                libXcursor
                libXfixes
                libXi
                libXext
                libXrender
                xorgproto
              ])
              ++ [
                libGL
                libpng
                wayland
                libxkbcommon
                wayland-protocols
              ];
          };
        }
      );

      # Модуль NixOS: programs.afk-fortnite.enable = true;
      nixosModules.default = import ./nixos-module.nix self;
    };
}
