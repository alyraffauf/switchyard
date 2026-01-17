{inputs, ...}: {
  perSystem = {
    pkgs,
    self',
    ...
  }: {
    packages = {
      default = self'.packages.switchyard;

      switchyard = pkgs.buildGoModule {
        pname = "switchyard";
        version = "dev";
        src = inputs.self;
        vendorHash = null;
        subPackages = ["src"];

        postBuild = ''
          mv $GOPATH/bin/src $GOPATH/bin/switchyard || true
        '';

        nativeBuildInputs = with pkgs; [
          pkg-config
          wrapGAppsHook4
        ];

        buildInputs = with pkgs; [
          glib
          gobject-introspection
          gtk4
          libadwaita
          pkg-config
        ];

        postInstall = ''
          install -Dm644 data/io.github.alyraffauf.Switchyard.desktop \
            $out/share/applications/io.github.alyraffauf.Switchyard.desktop

          install -Dm644 data/icons/hicolor/scalable/apps/io.github.alyraffauf.Switchyard.svg \
            $out/share/icons/hicolor/scalable/apps/io.github.alyraffauf.Switchyard.svg
          install -Dm644 data/icons/hicolor/symbolic/apps/io.github.alyraffauf.Switchyard-symbolic.svg \
            $out/share/icons/hicolor/symbolic/apps/io.github.alyraffauf.Switchyard-symbolic.svg
          install -Dm644 data/icons/hicolor/128x128/apps/io.github.alyraffauf.Switchyard.png \
            $out/share/icons/hicolor/128x128/apps/io.github.alyraffauf.Switchyard.png
          install -Dm644 data/icons/hicolor/64x64/apps/io.github.alyraffauf.Switchyard.png \
            $out/share/icons/hicolor/64x64/apps/io.github.alyraffauf.Switchyard.png
        '';

        meta = with pkgs.lib; {
          description = "A configurable default browser for Linux";
          homepage = "https://github.com/alyraffauf/switchyard";
          license = licenses.gpl3Plus;
          platforms = platforms.linux;
          mainProgram = "switchyard";
        };
      };
    };
  };
}
