{inputs, ...}: {
  flake.homeManagerModules.switchyard = import ./hm-module.nix inputs.self;

  perSystem = {
    pkgs,
    self',
    ...
  }: {
    packages = {
      default = self'.packages.switchyard;

      browser-setup = pkgs.python3Packages.buildPythonApplication {
        pname = "browser-setup";
        version = "dev";
        src = ../browser-setup;
        pyproject = true;
        build-system = with pkgs.python3.pkgs; [ setuptools ];
      };

      switchyard = pkgs.buildGoModule {
        pname = "switchyard";
        version = "dev";
        src = inputs.self;
        vendorHash = "sha256-1rEtf5QJA5xaeL0LRp13dwtPCH+DM5Wp4kJXBBqNtEg=";
        subPackages = ["src"];

        buildPhase = ''
          runHook preBuild
          go build -trimpath -ldflags="-s -w" -o $GOPATH/bin/switchyard ./src
          runHook postBuild
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
