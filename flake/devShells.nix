_: {
  perSystem = {
    config,
    lib,
    pkgs,
    ...
  }: {
    devShells.default = pkgs.mkShell {
      packages =
        (with pkgs; [
          glib
          gobject-introspection
          gtk4
          just
          libadwaita
          pkg-config
        ])
        ++ lib.attrValues config.treefmt.build.programs;

      shellHook = ''
        echo "Installing pre-commit hooks..."
        ${config.pre-commit.installationScript}
        export FLAKE="." NH_FLAKE="."
        echo "👋 Welcome to the switchyard devShell!"
      '';
    };
  };
}
