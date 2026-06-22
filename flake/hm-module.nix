self: {
  pkgs,
  config,
  lib,
  ...
}: let
  cfg = config.programs.switchyard;
  toml = pkgs.formats.toml {};
in {
  options.programs.switchyard = {
    enable = lib.mkEnableOption "Switchyard browser launcher";

    package = lib.mkPackageOption self.packages.${pkgs.system} "switchyard" {};

    setAsDefaultBrowser = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = ''
        Register switchyard as the default handler for http(s) and text/html via {manpage}`home-configuration.nix(5)`'s {option}`xdg.mimeApps`.
      '';
    };

    settings = lib.mkOption {
      inherit (toml) type;
      default = {};
      description = ''
        Contents of {file}`$XDG_CONFIG_HOME/switchyard/config.toml`.

        Omitted keys fall back to switchyard's built-in defaults. Note that Switchyard overwrites this file on every in-app config change, so GUI edits will be lost on the next {command}`home-manager switch`.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [cfg.package];

    xdg.configFile."switchyard/config.toml" = lib.mkIf (cfg.settings != {}) {
      source = toml.generate "switchyard-config.toml" cfg.settings;
    };

    xdg.mimeApps.defaultApplications = lib.mkIf cfg.setAsDefaultBrowser {
      "x-scheme-handler/http" = ["io.github.alyraffauf.Switchyard.desktop"];
      "x-scheme-handler/https" = ["io.github.alyraffauf.Switchyard.desktop"];
      "text/html" = ["io.github.alyraffauf.Switchyard.desktop"];
    };
  };
}
