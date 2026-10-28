{inputs, ...}: {
  flake.homeManagerModules.switchyard = import ./hm-module.nix inputs.self;

  perSystem = {
    goToolchain,
    pkgs,
    self',
    ...
  }: let
    buildGoModule = pkgs.buildGo127Module.override {go = goToolchain;};
    # Shared across both Go packages
    vendorHash = "sha256-pGON6Eq9P+oXk3LE6V1P11KMaAsagjs2e7Lc27a36QM=";
  in {
    packages =
      {
        # Terminal browser picker (cmd/sw). The only package that builds on
        # both Linux and macOS.
        sw = pkgs.callPackage ../nix/sw.nix {inherit buildGoModule vendorHash;};
      }
      // pkgs.lib.optionalAttrs pkgs.stdenv.hostPlatform.isLinux {
        default = self'.packages.switchyard;

        # Flatpak/XDG desktop integration installer -- Linux-only concepts.
        browser-setup = pkgs.python3Packages.buildPythonApplication {
          pname = "browser-setup";
          version = "dev";
          src = ../browser-setup;
          pyproject = true;
          build-system = with pkgs.python3.pkgs; [setuptools];
        };

        # GTK GUI (cmd/switchyard), Linux-only.
        switchyard = pkgs.callPackage ../nix/switchyard.nix {inherit buildGoModule vendorHash;};
      }
      // pkgs.lib.optionalAttrs pkgs.stdenv.hostPlatform.isDarwin {
        default = self'.packages.sw;
      };
  };
}
