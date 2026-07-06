{
  buildGoModule,
  glib,
  gobject-introspection,
  gtk4,
  libadwaita,
  lib,
  pkg-config,
  vendorHash,
  wrapGAppsHook4,
}:
buildGoModule {
  pname = "switchyard";
  version = "dev";
  src = ../.;
  inherit vendorHash;
  subPackages = ["cmd/switchyard"];

  ldflags = [
    "-s"
    "-w"
  ];

  nativeBuildInputs = [
    pkg-config
    wrapGAppsHook4
  ];

  buildInputs = [
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

  meta = with lib; {
    description = "A configurable default browser for Linux";
    homepage = "https://github.com/alyraffauf/switchyard";
    license = licenses.gpl3Plus;
    platforms = platforms.linux;
    mainProgram = "switchyard";
  };
}
